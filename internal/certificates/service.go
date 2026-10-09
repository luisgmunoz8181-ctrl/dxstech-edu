package certificates

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
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

	"dxstech-edu/internal/audit"
	"dxstech-edu/internal/auth"
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
	ID             string    `json:"id"`
	UserID         string    `json:"userId,omitempty"`
	CourseID       string    `json:"courseId,omitempty"`
	StudentName    string    `json:"studentName"`
	CourseTitle    string    `json:"courseTitle"`
	DurationHours  float64   `json:"durationHours"`
	InstructorName string    `json:"instructorName"`
	IssueDate      string    `json:"issueDate"`
	QRCodeURL      string    `json:"qrCodeUrl"`
	CreatedAt      time.Time `json:"createdAt"`
}

// PublicCertificate es la vista pública (sin autenticación) de un certificado.
type PublicCertificate struct {
	ID             string  `json:"id"`
	StudentName    string  `json:"studentName"`
	CourseTitle    string  `json:"courseTitle"`
	DurationHours  float64 `json:"durationHours"`
	InstructorName string  `json:"instructorName"`
	IssueDate      string  `json:"issueDate"`
}

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

func (s *Service) RegisterRoutes(r *gin.RouterGroup) {
	// Generador Masivo (Retrocompatibilidad total)
	admin := auth.RequireRole("SUPERADMIN", "ADMINISTRADOR")
	r.POST("/generate", admin, s.GenerateCertificates)
	r.GET("/issued", admin, s.ListIssuedCertificates)

	// Verificación pública (la consulta por código QR no requiere sesión)
	// Los códigos inexistentes cuentan para un límite por IP: evita enumerar certificados.
	verifyLimiter := auth.NewRateLimiter(30, time.Minute)
	r.GET("/verify/:id", verifyLimiter.StatusFailureMiddleware(http.StatusNotFound), s.VerifyCertificate)

	// Endpoints Oficiales LMS Fase 4
	r.GET("/my-certificates", auth.RequireAuth(), s.MyCertificates)
	r.GET("/course/:courseId", auth.RequireAuth(), s.GetCourseCertificate)
	// El PDF contiene datos personales: solo el titular o un administrador.
	r.GET("/:id/pdf", auth.RequireAuth(), s.DownloadCertificatePDF)
	r.GET("/download/:id", auth.RequireAuth(), s.DownloadCertificatePDF)
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
	issueDateStr := formatSpanishDate(time.Now())

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

				qrX := (req.QRXPercent/100.0)*pageWd - (req.QRSize / 2.0)
				qrY := (req.QRYPercent/100.0)*pageHt - (req.QRSize / 2.0)
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
						INSERT INTO issued_certificates (id, student_name, course_title, duration_hours, instructor_name, issue_date, qr_code_url, created_at)
						VALUES (?, ?, ?, 10.0, 'DxSTech Edu', ?, ?, CURRENT_TIMESTAMP)
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

	audit.Record(s.db, c, audit.CertGenerateBulk, "issued_certificates", "Generación masiva: %d certificados — %s", len(validStudents), req.CourseTitle)
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

	cert, err := s.GetCertificateByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"valid": false,
			"error": "El certificado no fue encontrado en la base de datos oficial de DxSTech Edu. Verifique el código ingresado.",
		})
		return
	}

	// Respuesta pública mínima: lo necesario para validar el diploma, sin
	// identificadores internos de usuario ni de curso.
	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"certificate": PublicCertificate{
			ID:             cert.ID,
			StudentName:    cert.StudentName,
			CourseTitle:    cert.CourseTitle,
			DurationHours:  cert.DurationHours,
			InstructorName: cert.InstructorName,
			IssueDate:      cert.IssueDate,
		},
		"issuer":      "DxSTech Edu — Academy of Technology & AI",
		"status":      "Certificado Oficial Verificado",
	})
}

func (s *Service) ListIssuedCertificates(c *gin.Context) {
	rows, err := s.db.QueryContext(c.Request.Context(), `
		SELECT id, user_id, course_id, student_name, course_title, duration_hours, instructor_name, issue_date, qr_code_url, created_at
		FROM issued_certificates
		ORDER BY created_at DESC
		LIMIT 100
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error consultando certificados emitidos"})
		return
	}
	defer rows.Close()

	certs := make([]IssuedCertificate, 0)
	for rows.Next() {
		var cert IssuedCertificate
		var uID, cID, instName sql.NullString
		var durationHours sql.NullFloat64
		if err := rows.Scan(&cert.ID, &uID, &cID, &cert.StudentName, &cert.CourseTitle, &durationHours, &instName, &cert.IssueDate, &cert.QRCodeURL, &cert.CreatedAt); err == nil {
			if uID.Valid {
				cert.UserID = uID.String
			}
			if cID.Valid {
				cert.CourseID = cID.String
			}
			if durationHours.Valid {
				cert.DurationHours = durationHours.Float64
			}
			if instName.Valid {
				cert.InstructorName = instName.String
			} else {
				cert.InstructorName = "Dirección Académica DxSTech"
			}
			certs = append(certs, cert)
		}
	}

	c.JSON(http.StatusOK, certs)
}

func (s *Service) MyCertificates(c *gin.Context) {
	userID := c.GetString("userID")
	if strings.TrimSpace(userID) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Debe iniciar sesión para ver sus certificados"})
		return
	}

	certs, err := s.ListUserCertificates(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error consultando sus certificados: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, certs)
}

func (s *Service) GetCourseCertificate(c *gin.Context) {
	courseID := strings.TrimSpace(c.Param("courseId"))
	userID := c.GetString("userID")
	if strings.TrimSpace(userID) == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Debe iniciar sesión para ver su certificado"})
		return
	}

	cert, err := s.GetCertificateByUserAndCourse(c.Request.Context(), userID, courseID)
	if err != nil {
		// Validar si la matrícula ya alcanzó el 100% y autogenerar inmediatamente
		var progressPercent float64
		var status string
		enrErr := s.db.QueryRowContext(c.Request.Context(), "SELECT progress_percent, status FROM enrollments WHERE user_id = ? AND course_id = ?", userID, courseID).Scan(&progressPercent, &status)
		if enrErr == nil && progressPercent >= 100.0 {
			var firstName, lastName, courseTitle, instructorName string
			var durationHours float64
			_ = s.db.QueryRowContext(c.Request.Context(), "SELECT first_name, last_name FROM users WHERE id = ?", userID).Scan(&firstName, &lastName)
			_ = s.db.QueryRowContext(c.Request.Context(), "SELECT title, duration_hours, instructor_name FROM courses WHERE id = ?", courseID).Scan(&courseTitle, &durationHours, &instructorName)

			studentName := strings.TrimSpace(firstName + " " + lastName)
			if studentName == "" {
				studentName = "Estudiante DxSTech"
			}
			if instructorName == "" {
				instructorName = "Dirección Académica DxSTech"
			}

			scheme := "https"
			if c.Request.TLS == nil && c.GetHeader("X-Forwarded-Proto") != "https" && strings.HasPrefix(c.Request.Host, "localhost") {
				scheme = "http"
			}
			baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

			newCert, issueErr := s.IssueCourseCertificate(c.Request.Context(), userID, courseID, studentName, courseTitle, durationHours, instructorName, baseURL)
			if issueErr == nil {
				c.JSON(http.StatusOK, newCert)
				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "Aún no has completado el 100% de este curso para obtener tu certificado oficial."})
		return
	}

	c.JSON(http.StatusOK, cert)
}

func (s *Service) DownloadCertificatePDF(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Código de certificado requerido"})
		return
	}

	cert, err := s.GetCertificateByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificado no encontrado en la base oficial"})
		return
	}

	role := c.GetString("userRole")
	isAdmin := role == "SUPERADMIN" || role == "ADMINISTRADOR"
	if !isAdmin && (cert.UserID == "" || cert.UserID != c.GetString("userID")) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Solo el titular del certificado o un administrador puede descargar el PDF"})
		return
	}

	pdfBytes, err := s.GenerateOfficialPDF(cert)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generando PDF oficial: " + err.Error()})
		return
	}

	sanitized := sanitizeFilename(cert.StudentName)
	filename := fmt.Sprintf("Certificado_DxSTech_%s_%s.pdf", cert.ID, sanitized)

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
	c.Header("Cache-Control", "private, max-age=3600")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func (s *Service) IssueCourseCertificate(ctx context.Context, userID, courseID, studentName, courseTitle string, durationHours float64, instructorName, baseURL string) (*IssuedCertificate, error) {
	// Verificar si ya existe
	existing, err := s.GetCertificateByUserAndCourse(ctx, userID, courseID)
	if err == nil && existing != nil {
		return existing, nil
	}

	certID := generateCertID()
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://dxstech-edu.onrender.com"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	verificationURL := fmt.Sprintf("%s/#verify/%s", baseURL, certID)

	issueDateStr := formatSpanishDate(time.Now())

	if strings.TrimSpace(instructorName) == "" {
		instructorName = "Dirección Académica DxSTech"
	}
	if durationHours <= 0 {
		durationHours = 10.0
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO issued_certificates (id, user_id, course_id, student_name, course_title, duration_hours, instructor_name, issue_date, qr_code_url, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, certID, userID, courseID, studentName, courseTitle, durationHours, instructorName, issueDateStr, verificationURL)
	if err != nil {
		return nil, fmt.Errorf("error guardando certificado emitido: %w", err)
	}
	audit.Log(s.db, audit.Entry{UserID: userID, Action: audit.CertIssued, Resource: "issued_certificates",
		Details: fmt.Sprintf("Certificado %s emitido a %s por el curso %s", certID, studentName, courseTitle)})

	return &IssuedCertificate{
		ID:             certID,
		UserID:         userID,
		CourseID:       courseID,
		StudentName:    studentName,
		CourseTitle:    courseTitle,
		DurationHours:  durationHours,
		InstructorName: instructorName,
		IssueDate:      issueDateStr,
		QRCodeURL:      verificationURL,
		CreatedAt:      time.Now(),
	}, nil
}

func (s *Service) GetCertificateByID(ctx context.Context, id string) (*IssuedCertificate, error) {
	var cert IssuedCertificate
	var userID, courseID, instructorName sql.NullString
	var durationHours sql.NullFloat64

	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, course_id, student_name, course_title, duration_hours, instructor_name, issue_date, qr_code_url, created_at
		FROM issued_certificates
		WHERE id = ?
	`, id).Scan(&cert.ID, &userID, &courseID, &cert.StudentName, &cert.CourseTitle, &durationHours, &instructorName, &cert.IssueDate, &cert.QRCodeURL, &cert.CreatedAt)
	if err != nil {
		return nil, err
	}

	if userID.Valid {
		cert.UserID = userID.String
	}
	if courseID.Valid {
		cert.CourseID = courseID.String
	}
	if durationHours.Valid {
		cert.DurationHours = durationHours.Float64
	}
	if instructorName.Valid {
		cert.InstructorName = instructorName.String
	} else {
		cert.InstructorName = "Dirección Académica DxSTech"
	}
	return &cert, nil
}

func (s *Service) GetCertificateByUserAndCourse(ctx context.Context, userID, courseID string) (*IssuedCertificate, error) {
	var cert IssuedCertificate
	var uID, cID, instructorName sql.NullString
	var durationHours sql.NullFloat64

	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, course_id, student_name, course_title, duration_hours, instructor_name, issue_date, qr_code_url, created_at
		FROM issued_certificates
		WHERE user_id = ? AND course_id = ?
		LIMIT 1
	`, userID, courseID).Scan(&cert.ID, &uID, &cID, &cert.StudentName, &cert.CourseTitle, &durationHours, &instructorName, &cert.IssueDate, &cert.QRCodeURL, &cert.CreatedAt)
	if err != nil {
		return nil, err
	}

	if uID.Valid {
		cert.UserID = uID.String
	}
	if cID.Valid {
		cert.CourseID = cID.String
	}
	if durationHours.Valid {
		cert.DurationHours = durationHours.Float64
	}
	if instructorName.Valid {
		cert.InstructorName = instructorName.String
	} else {
		cert.InstructorName = "Dirección Académica DxSTech"
	}
	return &cert, nil
}

func (s *Service) ListUserCertificates(ctx context.Context, userID string) ([]IssuedCertificate, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, course_id, student_name, course_title, duration_hours, instructor_name, issue_date, qr_code_url, created_at
		FROM issued_certificates
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var certs []IssuedCertificate
	for rows.Next() {
		var cert IssuedCertificate
		var uID, cID, instructorName sql.NullString
		var durationHours sql.NullFloat64
		if err := rows.Scan(&cert.ID, &uID, &cID, &cert.StudentName, &cert.CourseTitle, &durationHours, &instructorName, &cert.IssueDate, &cert.QRCodeURL, &cert.CreatedAt); err == nil {
			if uID.Valid {
				cert.UserID = uID.String
			}
			if cID.Valid {
				cert.CourseID = cID.String
			}
			if durationHours.Valid {
				cert.DurationHours = durationHours.Float64
			}
			if instructorName.Valid {
				cert.InstructorName = instructorName.String
			} else {
				cert.InstructorName = "Dirección Académica DxSTech"
			}
			certs = append(certs, cert)
		}
	}
	if certs == nil {
		certs = []IssuedCertificate{}
	}
	return certs, nil
}

func (s *Service) GenerateOfficialPDF(cert *IssuedCertificate) ([]byte, error) {
	pageWd := 297.0 // mm (A4 landscape)
	pageHt := 210.0

	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		OrientationStr: "L",
		UnitStr:        "mm",
		Size:           gofpdf.SizeType{Wd: pageWd, Ht: pageHt},
	})
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// 1. Fondo suave crema cálido
	pdf.SetFillColor(253, 252, 248)
	pdf.Rect(0, 0, pageWd, pageHt, "F")

	// 2. Borde exterior elegante - Azul Marino Real (#0f172a)
	pdf.SetDrawColor(15, 23, 42)
	pdf.SetLineWidth(2.2)
	pdf.Rect(10, 10, pageWd-20, pageHt-20, "D")

	// 3. Borde interior sutil - Dorado (#d97706)
	pdf.SetDrawColor(217, 119, 6)
	pdf.SetLineWidth(0.8)
	pdf.Rect(14, 14, pageWd-28, pageHt-28, "D")

	// 4. Esquineros ornamentales dorados
	pdf.SetFillColor(217, 119, 6)
	pdf.Rect(12.5, 12.5, 3.0, 3.0, "FD")
	pdf.Rect(pageWd-15.5, 12.5, 3.0, 3.0, "FD")
	pdf.Rect(12.5, pageHt-15.5, 3.0, 3.0, "FD")
	pdf.Rect(pageWd-15.5, pageHt-15.5, 3.0, 3.0, "FD")

	centerX := pageWd / 2.0

	// 5. Encabezado Institucional: DXSTECH EDU
	pdf.SetFont("Helvetica", "B", 22)
	pdf.SetTextColor(15, 23, 42)
	instName := tr("DXSTECH EDU")
	pdf.Text(centerX-(pdf.GetStringWidth(instName)/2.0), 28, instName)

	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(100, 116, 139)
	instSub := tr("ACADEMIA DE FORMACIÓN TECNOLÓGICA & INTELIGENCIA ARTIFICIAL")
	pdf.Text(centerX-(pdf.GetStringWidth(instSub)/2.0), 34, instSub)

	// Divisor decorativo con acento dorado
	pdf.SetDrawColor(217, 119, 6)
	pdf.SetLineWidth(0.5)
	pdf.Line(centerX-50, 39, centerX+50, 39)
	pdf.Circle(centerX, 39, 1.2, "FD")

	// 6. Título del Documento
	pdf.SetFont("Helvetica", "B", 20)
	pdf.SetTextColor(30, 41, 59)
	docTitle := tr("CERTIFICADO DE FINALIZACIÓN Y APROBACIÓN")
	pdf.Text(centerX-(pdf.GetStringWidth(docTitle)/2.0), 52, docTitle)

	// 7. Frase de Presentación
	pdf.SetFont("Times", "I", 12)
	pdf.SetTextColor(71, 85, 105)
	recog := tr("Se otorga el presente reconocimiento oficial y acreditación a:")
	pdf.Text(centerX-(pdf.GetStringWidth(recog)/2.0), 64, recog)

	// 8. Nombre del Estudiante (Grande, elegante)
	pdf.SetFont("Helvetica", "B", 26)
	pdf.SetTextColor(15, 23, 42)
	stName := tr(strings.ToUpper(cert.StudentName))
	pdf.Text(centerX-(pdf.GetStringWidth(stName)/2.0), 80, stName)

	// Línea dorada de subrayado
	stW := pdf.GetStringWidth(stName)
	if stW < 80 {
		stW = 80
	}
	pdf.SetDrawColor(217, 119, 6)
	pdf.SetLineWidth(0.8)
	pdf.Line(centerX-(stW/2.0), 84, centerX+(stW/2.0), 84)

	// 9. Frase de Aprobación
	pdf.SetFont("Times", "I", 11)
	pdf.SetTextColor(71, 85, 105)
	reason := tr("Por haber culminado con éxito todos los módulos, evaluaciones prácticas y requisitos del programa:")
	pdf.Text(centerX-(pdf.GetStringWidth(reason)/2.0), 96, reason)

	// 10. Título del Curso
	pdf.SetFont("Helvetica", "B", 18)
	pdf.SetTextColor(30, 58, 138) // Azul real (#1e3a8a)
	cTitle := tr(fmt.Sprintf("« %s »", cert.CourseTitle))
	pdf.Text(centerX-(pdf.GetStringWidth(cTitle)/2.0), 110, cTitle)

	// 11. Métricas del Curso
	hours := cert.DurationHours
	if hours <= 0 {
		hours = 12.0
	}
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(51, 65, 85)
	detailsStr := tr(fmt.Sprintf("Intensidad Académica: %.1f Horas   •   Modalidad: Virtual Asincrónica   •   Fecha: %s", hours, cert.IssueDate))
	pdf.Text(centerX-(pdf.GetStringWidth(detailsStr)/2.0), 122, detailsStr)

	// Divisor inferior
	pdf.SetDrawColor(226, 232, 240)
	pdf.SetLineWidth(0.4)
	pdf.Line(25, 138, pageWd-25, 138)

	// 12. Bloque de Firmas y Validación (Fila inferior)
	// Columna Izquierda: Firma del Instructor / Director
	instructor := cert.InstructorName
	if strings.TrimSpace(instructor) == "" {
		instructor = "Ing. Carlos Mendoza"
	}
	pdf.SetDrawColor(15, 23, 42)
	pdf.SetLineWidth(0.6)
	pdf.Line(32, 166, 102, 166)

	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(15, 23, 42)
	instTxt := tr(instructor)
	pdf.Text(32, 172, instTxt)

	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(100, 116, 139)
	pdf.Text(32, 177, tr("Director Académico & Certificación"))
	pdf.Text(32, 181, tr("DxSTech Edu"))

	// Columna Central: Sello Oficial Vectorial
	sealX := centerX
	sealY := 162.0
	pdf.SetDrawColor(217, 119, 6)
	pdf.SetLineWidth(0.8)
	pdf.SetFillColor(254, 243, 199) // Amber 100
	pdf.Circle(sealX, sealY, 14, "FD")
	pdf.Circle(sealX, sealY, 12, "D")

	pdf.SetFont("Helvetica", "B", 6)
	pdf.SetTextColor(15, 23, 42)
	sTop := tr("VALIDEZ OFICIAL")
	pdf.Text(sealX-(pdf.GetStringWidth(sTop)/2.0), sealY-4, sTop)

	pdf.SetFont("Helvetica", "B", 8)
	pdf.SetTextColor(180, 83, 9)
	sMid := tr("ACADÉMICA")
	pdf.Text(sealX-(pdf.GetStringWidth(sMid)/2.0), sealY, sMid)

	pdf.SetFont("Helvetica", "B", 5.5)
	pdf.SetTextColor(100, 116, 139)
	sBot := tr("DXSTECH EDU")
	pdf.Text(sealX-(pdf.GetStringWidth(sBot)/2.0), sealY+5, sBot)

	// Columna Derecha: Código QR Oficial & Código
	qrSize := 22.0
	qrX := pageWd - 75.0
	qrY := 146.0

	qrURL := cert.QRCodeURL
	if qrURL == "" {
		qrURL = fmt.Sprintf("https://dxstech-edu.onrender.com/#verify/%s", cert.ID)
	}

	qrPNG, qrErr := qrcode.Encode(qrURL, qrcode.Medium, 256)
	if qrErr == nil {
		qrImgName := fmt.Sprintf("qr_off_%s", cert.ID)
		pdf.RegisterImageOptionsReader(qrImgName, gofpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(qrPNG))
		pdf.ImageOptions(qrImgName, qrX, qrY, qrSize, qrSize, false, gofpdf.ImageOptions{ImageType: "PNG"}, 0, "")

		// Código y texto al lado o debajo del QR
		pdf.SetFont("Helvetica", "B", 7)
		pdf.SetTextColor(15, 23, 42)
		codeLbl := tr("Registro Oficial:")
		pdf.Text(qrX+qrSize+3.0, qrY+6.0, codeLbl)

		pdf.SetFont("Courier", "B", 8)
		pdf.SetTextColor(180, 83, 9)
		pdf.Text(qrX+qrSize+3.0, qrY+11.0, cert.ID)

		pdf.SetFont("Helvetica", "", 6)
		pdf.SetTextColor(100, 116, 139)
		pdf.Text(qrX+qrSize+3.0, qrY+16.0, tr("Escanee el código QR"))
		pdf.Text(qrX+qrSize+3.0, qrY+19.5, tr("para verificar autenticidad"))
	}

	// 13. Pie de Página de Seguridad y Verificación
	pdf.SetFont("Helvetica", "", 6)
	pdf.SetTextColor(148, 163, 184)
	footMsg := tr(fmt.Sprintf("Documento emitido electrónicamente por DxSTech Edu. Verifique su autenticidad en tiempo real en dxstech-edu.onrender.com/#verify/%s", cert.ID))
	pdf.Text(centerX-(pdf.GetStringWidth(footMsg)/2.0), 198, footMsg)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("error generando buffer PDF: %w", err)
	}

	return buf.Bytes(), nil
}

func generateCertID() string {
	// 48 bits de aleatoriedad: hacen inviable enumerar códigos válidos.
	bytes := make([]byte, 6)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("DXS-%d-%s", time.Now().Year(), strings.ToUpper(hex.EncodeToString(bytes)))
}

func formatSpanishDate(t time.Time) string {
	months := [...]string{
		"Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio",
		"Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre",
	}
	m := months[t.Month()-1]
	return fmt.Sprintf("%02d de %s de %d", t.Day(), m, t.Year())
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
