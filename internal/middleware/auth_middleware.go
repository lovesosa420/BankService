package middleware

import (
	"BankService/internal/domain/models"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func ValidateNParseJWT(c *fiber.Ctx) error {
	var request models.JWT
	if err := c.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	userToken := request.Token
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(userToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("secret"), nil
	})
	if err != nil || !token.Valid {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
	}

	if claims.Subject != "" {
		c.Locals("userID", claims.Subject)
	} else {
		c.Locals("refreshToken", userToken)
	}

	return c.Next()
}
