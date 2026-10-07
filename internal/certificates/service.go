package certificates

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"dxstech-edu/internal/database"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	qrcode "github.com/skip2/go-qrcode"
	"golang.org/x/image/webp"
)

type GenerateRequest struct {
	TemplateBase64 string   `json:"template" binding:"required"`
	XPercent       float64  `json:"xPercent"`
	YPercent       float64  `json:"yPercent"`
	FontSize       float64  `json:"fontSize"`
	FontColor      string   `json:"fontColor"`
	FontFamily     string   `json:"fontFamily"`
	TextAlign      string   `json:"textAlign"`
	IncludeQR      bool     `json:"includeQR"`
	QRXPercent     float64  `json:"qrXPercent"`
	QRYPercent     float64  `json:"qrYPercent"`
	QRSize         float64  `json:"qrSize"`
	CourseTitle    string   `json:"courseTitle"`
	Students       []string `json:"students" binding:"required"`
}

type IssuedCertificate struct {
	ID          string    `json:"id"`
	StudentName string    `json:"studentName"`
	CourseTitle string    `json:"courseTitle"`
	IssueDate   string    `json:"issueDate"`
	QRCodeURL   string    `json:"qrCodeUrl"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

func (s *Service) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/generate", s.GenerateCertificates)
	r.GET("/verify/:id", s.VerifyCertificate)
	r.GET("/issued", s.ListIssuedCertificates)
}

func (s *Service) GenerateCertificates(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parámetros inválidos para generación de certificados: " + err.Error()})
		return
	}

	if len(req.Students) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Debe proporcionar al menos un nombre de estudiante"})
		return
	}

	var validStudents []string
	for _, name := range req.Students {
		trimmed := strings.TrimSpace(name)
		if trimmed != "" {
			validStudents = append(validStudents, trimmed)
		}
	}
	if len(validStudents) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "La lista de estudiantes no contiene nombres válidos"})
		return
	}

	pngBytes, imgWidth, imgHeight, err := processTemplateImage(req.TemplateBase64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error al procesar la imagen de la plantilla: " + err.Error()})
		return
	}

	pageWd := 297.0 // mm (A4 landscape)
	pageHt := 210.0
	orientation := "L"
	if imgHeight > imgWidth {
		pageWd = 210.0
		pageHt = 297.0
		orientation = "P"
	}

	if req.FontSize <= 0 {
		req.FontSize = 32
	}
	if req.FontColor == "" {
		req.FontColor = "#1e293b"
	}
	fontFamily := "Helvetica"
	switch strings.ToLower(req.FontFamily) {
	case "times":
		fontFamily = "Times"
	case "courier":
		fontFamily = "Courier"
	case "arial", "helvetica", "sans":
		fontFamily = "Helvetica"
	}

	rVal, gVal, bVal := parseHexColor(req.FontColor)
	textAlign := strings.ToLower(req.TextAlign)
	if textAlign == "" {
		textAlign = "center"
	}

	// QR Defaults
	if req.QRSize <= 0 {
		req.QRSize = 22.0 // mm
	}
	if req.QRXPercent <= 0 {
		req.QRXPercent = 88.0
	}
	if req.QRYPercent <= 0 {
		req.QRYPercent = 82.0
	}
	if strings.TrimSpace(req.CourseTitle) == "" {
		req.CourseTitle = "Diplomado en Inteligencia Artificial y Desarrollo Web"
	}

	// Base URL for verification links
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	host := c.Request.Host
	if host == "" {
		host = "localhost:3000"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, host)

	// Create ZIP in memory
	zipBuf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuf)
	issueDateStr := time.Now().Format("02 de Enero de 2006")

	for i, studentName := range validStudents {
		pdf := gofpdf.NewCustom(&gofpdf.InitType{
			OrientationStr: orientation,
			UnitStr:        "mm",
			Size:           gofpdf.SizeType{Wd: pageWd, Ht: pageHt},
		})
		pdf.SetAutoPageBreak(false, 0)
		pdf.AddPage()

		// Embed template image
		imgName := fmt.Sprintf("tmpl_%d", i)
		pdf.RegisterImageOptionsReader(imgName, gofpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(pngBytes))
		pdf.ImageOptions(imgName, 0, 0, pageWd, pageHt, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")

		// Setup typography for student name
		pdf.SetFont(fontFamily, "B", req.FontSize)
		pdf.SetTextColor(rVal, gVal, bVal)

		// Calculate text positioning
		txtWidth := pdf.GetStringWidth(studentName)
		xPos := (req.XPercent / 100.0) * pageWd
		yPos := (req.YPercent / 100.0) * pageHt

		switch textAlign {
		case "center":
			xPos -= (txtWidth / 2.0)
		case "right":
			xPos -= txtWidth
		}

		pdf.Text(xPos, yPos, studentName)

		// Generate and embed Verification QR Code if requested
		if req.IncludeQR {
			certID := generateCertID()
			verificationURL := fmt.Sprintf("%s/#verify/%s", baseURL, certID)

			qrPNG, qrErr := qrcode.Encode(verificationURL, qrcode.Medium, 256)
			if qrErr == nil {
				qrImgName := fmt.Sprintf("qr_%d", i)
				pdf.RegisterImageOptionsReader(qrImgName, gofpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(qrPNG))

				qrX := (req.QRXPercent / 100.0) * pageWd - (req.QRSize / 2.0)
				qrY := (req.QRYPercent / 100.0) * pageHt - (req.QRSize / 2.0)
				pdf.ImageOptions(qrImgName, qrX, qrY, req.QRSize, req.QRSize, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")

				// Print small verification code below QR
				pdf.SetFont("Helvetica", "", 6)
				pdf.SetTextColor(100, 116, 139)
				codeStr := fmt.Sprintf("ID: %s", certID)
				codeW := pdf.GetStringWidth(codeStr)
				pdf.Text(qrX+(req.QRSize/2.0)-(codeW/2.0), qrY+req.QRSize+2.5, codeStr)

				// Save to SQLite
				if s.db != nil {
					_, _ = s.db.Exec(`
						INSERT INTO issued_certificates (id, student_name, course_title, issue_date, qr_code_url, created_at)
						VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
					`, certID, studentName, req.CourseTitle, issueDateStr, verificationURL)
				}
			}
		}

		var pdfBuf bytes.Buffer
		if err := pdf.Output(&pdfBuf); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error generando PDF para %s: %s", studentName, err.Error())})
			return
		}

		sanitizedName := sanitizeFilename(studentName)
		filename := fmt.Sprintf("Certificado_%s.pdf", sanitizedName)

		fileWriter, err := zipWriter.Create(filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al crear archivo en ZIP: " + err.Error()})
			return
		}

		if _, err := fileWriter.Write(pdfBuf.Bytes()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error escribiendo PDF en ZIP: " + err.Error()})
			return
		}
	}

	if err := zipWriter.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error cerrando archivo ZIP: " + err.Error()})
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=certificados_dxstech.zip")
	c.Data(http.StatusOK, "application/zip", zipBuf.Bytes())
}

func (s *Service) VerifyCertificate(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"valid": false, "error": "Código de certificado requerido"})
		return
	}

	var cert IssuedCertificate
	err := s.db.QueryRow(`
		SELECT id, student_name, course_title, issue_date, qr_code_url, created_at
		FROM issued_certificates
		WHERE id = ?
	`, id).Scan(&cert.ID, &cert.StudentName, &cert.CourseTitle, &cert.IssueDate, &cert.QRCodeURL, &cert.CreatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"valid": false,
			"error": "El certificado no fue encontrado en la base de datos oficial de DxSTech Edu. Verifique el código ingresado.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":       true,
		"certificate": cert,
		"issuer":      "DxSTech Edu — Academy of Technology & AI",
		"status":      "Certificado Oficial Verificado",
	})
}

func (s *Service) ListIssuedCertificates(c *gin.Context) {
	rows, err := s.db.Query(`
		SELECT id, student_name, course_title, issue_date, qr_code_url, created_at
		FROM issued_certificates
		ORDER BY id DESC
		LIMIT 50
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error consultando certificados emitidos"})
		return
	}
	defer rows.Close()

	certs := make([]IssuedCertificate, 0)
	for rows.Next() {
		var cert IssuedCertificate
		if err := rows.Scan(&cert.ID, &cert.StudentName, &cert.CourseTitle, &cert.IssueDate, &cert.QRCodeURL, &cert.CreatedAt); err == nil {
			certs = append(certs, cert)
		}
	}

	c.JSON(http.StatusOK, certs)
}

func generateCertID() string {
	bytes := make([]byte, 4)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("DXS-%d-%s", time.Now().Year(), strings.ToUpper(hex.EncodeToString(bytes)))
}

func processTemplateImage(base64Str string) ([]byte, int, int, error) {
	if idx := strings.Index(base64Str, ","); idx != -1 {
		base64Str = base64Str[idx+1:]
	}

	rawBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("error al decodificar base64: %w", err)
	}

	var img image.Image
	img, err = webp.Decode(bytes.NewReader(rawBytes))
	if err != nil {
		img, _, err = image.Decode(bytes.NewReader(rawBytes))
		if err != nil {
			return nil, 0, 0, errors.New("formato de imagen no soportado o archivo corrupto (se requiere WebP, PNG o JPEG)")
		}
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		return nil, 0, 0, fmt.Errorf("error convirtiendo imagen a PNG: %w", err)
	}

	return pngBuf.Bytes(), width, height, nil
}

func parseHexColor(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return 30, 41, 59
	}
	r, _ := strconv.ParseInt(hex[0:2], 16, 32)
	g, _ := strconv.ParseInt(hex[2:4], 16, 32)
	b, _ := strconv.ParseInt(hex[4:6], 16, 32)
	return int(r), int(g), int(b)
}

var invalidFileChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f\s]+`)

func sanitizeFilename(name string) string {
	clean := invalidFileChars.ReplaceAllString(name, "_")
	clean = strings.Trim(clean, "_")
	if clean == "" {
		clean = "Estudiante"
	}
	return clean
}
