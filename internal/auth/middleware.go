package auth

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"dxstech-edu/internal/database"

	"github.com/gin-gonic/gin"
)

// SessionState es el estado vigente de un usuario en base de datos.
type SessionState struct {
	Active             bool
	TokenVersion       int
	MustChangePassword bool
}

// SessionLookup consulta el estado actual de un usuario. Devuelve ok=false si
// el usuario no existe.
type SessionLookup func(ctx context.Context, userID string) (state SessionState, ok bool)

// NewSessionLookup crea un SessionLookup respaldado por la base de datos.
func NewSessionLookup(db *database.DB) SessionLookup {
	return func(ctx context.Context, userID string) (SessionState, bool) {
		var status string
		var tv, must int
		err := db.QueryRowContext(ctx,
			`SELECT status, token_version, must_change_password FROM users WHERE id = ?`, userID,
		).Scan(&status, &tv, &must)
		if err != nil {
			return SessionState{}, false
		}
		return SessionState{Active: status == "active", TokenVersion: tv, MustChangePassword: must == 1}, true
	}
}

// passwordChangeAllowedPaths son las únicas rutas disponibles mientras el
// usuario tenga must_change_password=1.
var passwordChangeAllowedPaths = map[string]bool{
	"/api/auth/me":              true,
	"/api/auth/change-password": true,
	"/api/auth/logout":          true,
}

const AuthCookieName = "dxstech_session"

func SetAuthCookie(c *gin.Context, token string, maxAgeSeconds int, isProd bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		AuthCookieName,
		token,
		maxAgeSeconds,
		"/",
		"",
		isProd, // Secure = true in production HTTPS
		true,   // HttpOnly = true (No readable by JS document.cookie)
	)
}

func ClearAuthCookie(c *gin.Context, isProd bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		AuthCookieName,
		"",
		-1,
		"/",
		"",
		isProd,
		true,
	)
}

// Authenticate extracts and validates token from cookie or Authorization header.
//
// Si se proporciona un SessionLookup, además de la firma se comprueba contra la
// base de datos que el usuario siga activo y que la versión del token no haya
// sido revocada (cambio de contraseña, de rol o desactivación).
func Authenticate(secret string, lookups ...SessionLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string

		// 1. Check HttpOnly cookie
		if cookie, err := c.Cookie(AuthCookieName); err == nil && strings.TrimSpace(cookie) != "" {
			tokenStr = cookie
		}

		// 2. Check Authorization header: Bearer <token>
		if tokenStr == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if strings.TrimSpace(tokenStr) == "" {
			c.Next()
			return
		}

		claims, err := ValidateToken(tokenStr, secret)
		if err == nil && claims != nil && len(lookups) > 0 && lookups[0] != nil {
			state, ok := lookups[0](c.Request.Context(), claims.UserID)
			if !ok || !state.Active || state.TokenVersion != claims.TokenVersion {
				claims = nil // sesión revocada: se trata como anónimo
			} else if state.MustChangePassword {
				c.Set("mustChangePassword", true)
			}
		}
		if err == nil && claims != nil {
			c.Set("user", claims)
			c.Set("userID", claims.UserID)
			c.Set("userRole", claims.Role)
			c.Set("userEmail", claims.Email)
			c.Set("userName", strings.TrimSpace(claims.FirstName+" "+claims.LastName))
		}

		c.Next()
	}
}

func abortIfPasswordChangeRequired(c *gin.Context) bool {
	if c.GetBool("mustChangePassword") && !passwordChangeAllowedPaths[c.Request.URL.Path] {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Debe cambiar su contraseña temporal antes de continuar",
			"code":  "PASSWORD_CHANGE_REQUIRED",
		})
		return true
	}
	return false
}

// RequireAuth enforces authenticated access
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsVal, exists := c.Get("user")
		if !exists || claimsVal == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Debe iniciar sesión para acceder a este recurso",
			})
			return
		}
		if abortIfPasswordChangeRequired(c) {
			return
		}
		c.Next()
	}
}

// RequireRole enforces specific roles (e.g., SUPERADMIN, ADMINISTRADOR)
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsVal, exists := c.Get("user")
		if !exists || claimsVal == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Debe iniciar sesión para acceder a este recurso",
			})
			return
		}

		claims, ok := claimsVal.(*Claims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Sesión inválida",
			})
			return
		}

		if abortIfPasswordChangeRequired(c) {
			return
		}

		for _, r := range allowedRoles {
			if strings.EqualFold(claims.Role, r) {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Acceso denegado: permisos insuficientes para realizar esta acción",
		})
	}
}

// MemoryRateLimiter implements IP-based rate limiting for sensitive auth routes
type MemoryRateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *MemoryRateLimiter {
	limiter := &MemoryRateLimiter{
		attempts: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Periodically purge old keys
	go func() {
		for {
			time.Sleep(2 * window)
			limiter.mu.Lock()
			now := time.Now()
			for ip, times := range limiter.attempts {
				var valid []time.Time
				for _, t := range times {
					if now.Sub(t) < window {
						valid = append(valid, t)
					}
				}
				if len(valid) == 0 {
					delete(limiter.attempts, ip)
				} else {
					limiter.attempts[ip] = valid
				}
			}
			limiter.mu.Unlock()
		}
	}()

	return limiter
}

// recent devuelve los intentos de la IP dentro de la ventana.
func (l *MemoryRateLimiter) recent(ip string, now time.Time) []time.Time {
	var valid []time.Time
	for _, t := range l.attempts[ip] {
		if now.Sub(t) < l.window {
			valid = append(valid, t)
		}
	}
	return valid
}

func tooManyAttempts(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"error": "Demasiados intentos de acceso fallidos. Por seguridad, intente de nuevo en unos minutos.",
	})
}

// Middleware cuenta TODAS las peticiones de la IP (útil para endpoints que
// siempre responden 200 por diseño, como forgot-password).
func (l *MemoryRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		l.mu.Lock()
		valid := l.recent(ip, now)
		if len(valid) >= l.limit {
			l.mu.Unlock()
			tooManyAttempts(c)
			return
		}
		l.attempts[ip] = append(valid, now)
		l.mu.Unlock()

		c.Next()
	}
}

// FailureMiddleware cuenta solo las peticiones que terminan en 401 (credenciales
// incorrectas). Así muchos accesos legítimos desde una misma IP (un aula, una
// oficina o una red con NAT) no se bloquean entre sí, mientras que quien
// adivina contraseñas sí queda limitado.
func (l *MemoryRateLimiter) FailureMiddleware() gin.HandlerFunc {
	return l.StatusFailureMiddleware(http.StatusUnauthorized)
}

// StatusFailureMiddleware cuenta solo las peticiones cuya respuesta tenga uno de
// los códigos indicados (p. ej. 404 en la verificación de certificados, para
// frenar la enumeración de códigos).
func (l *MemoryRateLimiter) StatusFailureMiddleware(codes ...int) gin.HandlerFunc {
	isFailure := func(status int) bool {
		for _, c := range codes {
			if c == status {
				return true
			}
		}
		return false
	}
	return func(c *gin.Context) {
		ip := c.ClientIP()

		l.mu.Lock()
		blocked := len(l.recent(ip, time.Now())) >= l.limit
		l.mu.Unlock()
		if blocked {
			tooManyAttempts(c)
			return
		}

		c.Next()

		if isFailure(c.Writer.Status()) {
			l.mu.Lock()
			now := time.Now()
			l.attempts[ip] = append(l.recent(ip, now), now)
			l.mu.Unlock()
		}
	}
}
