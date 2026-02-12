package middleware

import (
	"strings"

	"neighborhood-api/internal/services"
	"neighborhood-api/pkg/dto"
	"neighborhood-api/pkg/errors"
	"neighborhood-api/pkg/logger"

	"github.com/gin-gonic/gin"
)

func normalizeRole(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

// AuthMiddleware valida el JWT en las request
func AuthMiddleware(authService services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.Get()

		// Obtener token del header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Warn("Missing Authorization header")
			c.JSON(401, dto.ErrorResponse{
				Error:   true,
				Message: "missing authorization header",
				Code:    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		// Extraer token (format: "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Warn("Invalid Authorization header format")
			c.JSON(401, dto.ErrorResponse{
				Error:   true,
				Message: "invalid authorization header format",
				Code:    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Validar token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			if errors.Is(err, errors.UnauthorizedError) {
				log.WithError(err).Warn("Invalid token")
				c.JSON(401, dto.ErrorResponse{
					Error:   true,
					Message: "invalid token",
					Code:    "UNAUTHORIZED",
				})
			} else {
				log.WithError(err).Error("Token validation error")
				c.JSON(500, dto.ErrorResponse{
					Error:   true,
					Message: "internal server error",
					Code:    "INTERNAL_ERROR",
				})
			}
			c.Abort()
			return
		}

		// Guardar claims en contexto
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("condominio_id", claims.CondominioID)
		c.Set("role", normalizeRole(claims.Role))

		c.Next()
	}
}

// RequireRoles permite acceso solo a los roles indicados
func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		normalized := normalizeRole(role)
		if normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		log := logger.Get()

		roleValue, exists := c.Get("role")
		if !exists {
			log.Warn("Missing role in context")
			c.JSON(403, dto.ErrorResponse{
				Error:   true,
				Message: "forbidden",
				Code:    "FORBIDDEN",
			})
			c.Abort()
			return
		}

		role, ok := roleValue.(string)
		if !ok {
			log.Warn("Invalid role type in context")
			c.JSON(403, dto.ErrorResponse{
				Error:   true,
				Message: "forbidden",
				Code:    "FORBIDDEN",
			})
			c.Abort()
			return
		}

		if _, isAllowed := allowed[normalizeRole(role)]; !isAllowed {
			log.WithField("role", role).Warn("Forbidden role for endpoint")
			c.JSON(403, dto.ErrorResponse{
				Error:   true,
				Message: "forbidden",
				Code:    "FORBIDDEN",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireCondominio middleware que verifica que el condominio en params coincida
func RequireCondominio() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.Get()

		// Obtener condominio del contexto (establecido por AuthMiddleware)
		contextCondominioID, exists := c.Get("condominio_id")
		if !exists {
			log.Warn("Missing condominio_id in context")
			c.JSON(401, dto.ErrorResponse{
				Error:   true,
				Message: "unauthorized",
				Code:    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		// Obtener condominio del parámetro si existe
		paramCondominioID := c.Param("condominio_id")
		if paramCondominioID != "" && paramCondominioID != contextCondominioID {
			log.WithField("expected", contextCondominioID).
				WithField("provided", paramCondominioID).
				Warn("Condominio mismatch")

			c.JSON(403, dto.ErrorResponse{
				Error:   true,
				Message: "forbidden",
				Code:    "FORBIDDEN",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
