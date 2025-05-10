package main

import (
	"BankService/internal/domain/services"
	"BankService/internal/handlers"
	"BankService/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	handler := handlers.NewHandler()
	defer handler.Storage.Close()
	defer handler.JWT.Close()

	go services.RunScheduler()

	webApp := fiber.New()
	webApp.Post("/register", handler.Register)
	webApp.Post("/login", handler.Login)

	auth := webApp.Group("/bank", middleware.ValidateNParseJWT)
	auth.Get("/logout", handler.Logout)
	auth.Post("/account", handler.CreateAccount)
	auth.Post("/balance", handler.UpdateAccountBalance)
	auth.Post("/transfer", handler.TransferBetweenAccounts)
	auth.Post("/card", handler.CreateCard)
	auth.Get("/cards", handler.GetCards)
	auth.Post("/credit", handler.CreateCredit)

	logrus.Fatal(webApp.Listen(":8080"))
}
