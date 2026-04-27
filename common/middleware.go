package common

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// AuthRequired adalah middleware untuk memvalidasi Token JWT
func AuthRequired(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Ambil token dari header Authorization: Bearer <token>
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing or empty authorization header",
			})
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// 2. Parse dan Validasi Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// 3. Cek Role (Jika dibutuhkan)
		claims := token.Claims.(jwt.MapClaims)
		userRole := claims["role"].(string)

		// Jika role yang diminta adalah admin, tapi user adalah kasir, tolak!
		if role != "" && userRole != role {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: You don't have access to this resource",
			})
		}

		// Simpan data user ke context agar bisa dipakai di controller (misal: ID Kasir)
		c.Locals("user_id", claims["user_id"])
		c.Locals("role", userRole)

		return c.Next()
	}
}