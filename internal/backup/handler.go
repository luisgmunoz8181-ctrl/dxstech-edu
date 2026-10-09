package backup

import (
	"net/http"

	"dxstech-edu/internal/audit"
	"dxstech-edu/internal/auth"
	"dxstech-edu/internal/database"

	"github.com/gin-gonic/gin"
)

// Handler expone la gestión de copias de seguridad. Solo SUPERADMIN: la base de
// datos contiene hashes de contraseñas y datos personales de todos los usuarios.
type Handler struct {
	svc *Service
	db  *database.DB
}

func NewHandler(svc *Service, db *database.DB) *Handler { return &Handler{svc: svc, db: db} }

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.Use(auth.RequireAuth(), auth.RequireRole("SUPERADMIN"))
	r.GET("", h.List)
	r.POST("", h.Create)
	r.GET("/:name", h.Download)
}

func (h *Handler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudieron listar las copias de seguridad"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"backups": list, "retention": h.svc.retention})
}

func (h *Handler) Create(c *gin.Context) {
	info, err := h.svc.Create(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear la copia de seguridad"})
		return
	}
	audit.Record(h.db, c, audit.BackupCreate, "backups", "Copia de seguridad creada: %s (%d bytes)", info.Name, info.SizeBytes)
	c.JSON(http.StatusCreated, info)
}

func (h *Handler) Download(c *gin.Context) {
	path, err := h.svc.Path(c.Param("name"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	audit.Record(h.db, c, audit.BackupDownload, "backups", "Copia de seguridad descargada: %s", c.Param("name"))
	c.Header("Content-Disposition", `attachment; filename="`+c.Param("name")+`"`)
	c.Header("Cache-Control", "no-store")
	c.File(path)
}
