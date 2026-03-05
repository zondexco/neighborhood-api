package middleware

import (
	"neighborhood-api/internal/config"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware maneja CORS
func CORSMiddleware(corsConfig config.CORSConfig) gin.HandlerFunc {
	// Pre-compute allowed origins at startup instead of parsing on every request.
	allowedSet, wildcardPrefixes := buildOriginSets(corsConfig.AllowedOrigins)

	return func(c *gin.Context) {
		origin := normalizeOrigin(c.Request.Header.Get("Origin"))
		originAllowed := origin != "" && isOriginAllowedFast(origin, allowedSet, wildcardPrefixes)
		c.Writer.Header().Set("Vary", "Origin")

		// Validar que el origin esté permitido
		if originAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", strconv.FormatBool(corsConfig.AllowCredentials))
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", corsConfig.AllowedMethods)
		c.Writer.Header().Set("Access-Control-Allow-Headers", corsConfig.AllowedHeaders)
		c.Writer.Header().Set("Access-Control-Max-Age", strconv.Itoa(corsConfig.MaxAge))

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// buildOriginSets pre-computes a map for O(1) exact-match and a slice for
// wildcard-port origins (e.g. "http://localhost:*").
func buildOriginSets(rawOrigins string) (map[string]bool, []string) {
	rawOrigins = strings.TrimSpace(rawOrigins)
	exact := make(map[string]bool)
	var wildcards []string

	if rawOrigins == "*" {
		exact["*"] = true
		return exact, wildcards
	}

	for _, o := range strings.Split(rawOrigins, ",") {
		o = normalizeOrigin(strings.TrimSpace(o))
		if o == "" {
			continue
		}
		if strings.Contains(o, ":*") {
			wildcards = append(wildcards, strings.Replace(o, ":*", ":", 1))
		} else {
			exact[o] = true
		}
	}
	return exact, wildcards
}

// isOriginAllowedFast uses the pre-computed sets for fast lookups.
func isOriginAllowedFast(origin string, allowedSet map[string]bool, wildcardPrefixes []string) bool {
	if allowedSet["*"] {
		return true
	}
	if allowedSet[origin] {
		return true
	}
	for _, prefix := range wildcardPrefixes {
		if strings.HasPrefix(origin, prefix) {
			return true
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
