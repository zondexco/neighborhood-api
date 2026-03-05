package middleware

import (
	"neighborhood-api/internal/config"
	"neighborhood-api/pkg/logger"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware maneja CORS
func CORSMiddleware(corsConfig config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.Get()

		origin := normalizeOrigin(c.Request.Header.Get("Origin"))
		originAllowed := origin != "" && isOriginAllowed(origin, corsConfig.AllowedOrigins)
		c.Writer.Header().Set("Vary", "Origin")

		// Validar que el origin esté permitido
		if originAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", strconv.FormatBool(corsConfig.AllowCredentials))
		} else if origin != "" {
			log.WithField("origin", origin).
				WithField("allowed_origins", corsConfig.AllowedOrigins).
				WithField("method", c.Request.Method).
				Warn("CORS origin rejected")
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", corsConfig.AllowedMethods)
		c.Writer.Header().Set("Access-Control-Allow-Headers", corsConfig.AllowedHeaders)
		c.Writer.Header().Set("Access-Control-Max-Age", strconv.Itoa(corsConfig.MaxAge))

		log.WithField("origin", origin).
			WithField("allowed", originAllowed).
			WithField("method", c.Request.Method).
			Debug("CORS request processed")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isOriginAllowed verifica si un origin está en la lista de permitidos
func isOriginAllowed(origin string, allowedOrigins string) bool {
	origin = normalizeOrigin(origin)
	allowedOrigins = strings.TrimSpace(allowedOrigins)

	if origin == "" {
		return false
	}

	if allowedOrigins == "*" {
		return true
	}

	origins := strings.Split(allowedOrigins, ",")
	for _, o := range origins {
		o = normalizeOrigin(strings.TrimSpace(o))
		if o == "" {
			continue
		}

		// Exact match
		if o == origin {
			return true
		}
		// Wildcard match for localhost (e.g., "http://localhost:*")
		if strings.Contains(o, ":*") {
			baseOrigin := strings.Replace(o, ":*", "", 1)
			if strings.HasPrefix(origin, baseOrigin+":") {
				return true
			}
		}
	}

	return false
}

func normalizeOrigin(origin string) string {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return ""
	}

	parsed, err := url.Parse(origin)
	if err != nil {
		return strings.TrimSuffix(strings.ToLower(origin), "/")
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Path = ""
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return strings.TrimSuffix(parsed.String(), "/")
}
