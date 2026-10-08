package auth

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

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

// Authenticate extracts and validates token from cookie or Authorization header
func Authenticate(secret string) gin.HandlerFunc {
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
		if err == nil && claims != nil {
			c.Set("user", claims)
			c.Set("userID", claims.UserID)
			c.Set("userRole", claims.Role)
		}

		c.Next()
	}
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

func (l *MemoryRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		l.mu.Lock()
		times := l.attempts[ip]
		var valid []time.Time
		for _, t := range times {
			if now.Sub(t) < l.window {
				valid = append(valid, t)
			}
		}

		if len(valid) >= l.limit {
			l.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Demasiados intentos de acceso fallidos. Por seguridad, intente de nuevo en unos minutos.",
			})
			return
		}

		valid = append(valid, now)
		l.attempts[ip] = valid
		l.mu.Unlock()

		c.Next()
	}
}
