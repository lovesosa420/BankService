package handlers

import (
	"BankService/internal/domain/models"
	"BankService/internal/domain/services"
	"BankService/internal/repository"
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"sync"
)

type Handler struct {
	Storage *repository.Storage
	JWT     *repository.JWTStorage
	mu      sync.RWMutex
}

func NewHandler() *Handler {
	return &Handler{repository.NewStorage(), repository.NewJWTStorage(), sync.RWMutex{}}
}

func (h *Handler) HandleRefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Locals("refreshToken")
	if err := h.JWT.IsTokenExist(refreshToken.(string)); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	accessToken, err := h.JWT.GetAnotherToken(refreshToken.(string))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	claims := &jwt.RegisteredClaims{}
	_, err = jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("secret"), nil
	})
	user := claims.Subject
	if err = h.JWT.RemoveTokens(accessToken); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	newAccess, err := services.GenerateAccessToken(user)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	newRefresh, err := services.GenerateRefreshToken()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	err = h.JWT.SaveTokens(newAccess, newRefresh)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	response := models.UserLoginResponse{AccessToken: newAccess, RefreshToken: newRefresh}
	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var user models.UserRegisterRequest
	if err := c.BodyParser(&user); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)
	if err := h.Storage.RegisterUser(context.Background(), user); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.SendStatus(fiber.StatusOK)
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var user models.UserLoginRequest
	if err := c.BodyParser(&user); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	userData, err := h.Storage.GetUserData(context.Background(), user.Email)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "email or password is incorrect")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(userData.Password), []byte(user.Password)); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "email or password is incorrect")
	}

	accessToken, err := services.GenerateAccessToken(userData.Email)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	refreshToken, err := services.GenerateRefreshToken()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	err = h.JWT.SaveTokens(accessToken, refreshToken)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	response := models.UserLoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}
	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID != nil {
		accessToken := c.Locals("accessToken")
		if err := h.JWT.RemoveTokens(accessToken.(string)); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.Status(fiber.StatusOK).JSON(fmt.Sprintf("user with login %s was logout", userID.(string)))
	} else {
		return h.HandleRefreshToken(c)
	}
}

func (h *Handler) CreateAccount(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID != nil {
		var account models.Account
		if err := c.BodyParser(&account); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		account.Person = userID.(string)
		account.Balance = 0
		number, err := h.Storage.CreateAccount(context.Background(), account)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		response := models.AccountCreationResponse{Number: number, Type: account.Type}
		return c.Status(fiber.StatusOK).JSON(response)
	} else {
		return h.HandleRefreshToken(c)
	}
}

func (h *Handler) UpdateAccountBalance(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID != nil {
		var account models.AccountUpdateRequest
		if err := c.BodyParser(&account); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		account.Person = userID.(string)
		h.mu.RLock()
		balance, err := h.Storage.UpdateAccountBalance(context.Background(), account.Person, account.Number, account.Sum)
		h.mu.RUnlock()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.Status(fiber.StatusOK).JSON(fmt.Sprintf("current balance is %.2f", balance))
	} else {
		return h.HandleRefreshToken(c)
	}
}

func (h *Handler) TransferBetweenAccounts(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID != nil {
		var request models.AccountTransferRequest
		if err := c.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		request.Person = userID.(string)
		h.mu.RLock()
		balance, err := h.Storage.TransferBetweenAccounts(context.Background(), request.Person, request.From, request.To, request.Sum)
		h.mu.RUnlock()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.Status(fiber.StatusOK).JSON(fmt.Sprintf("current balance is %.2f", balance))
	} else {
		return h.HandleRefreshToken(c)
	}
}

func (h *Handler) CreateCard(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID != nil {
		var numAccount models.CardRequest
		if err := c.BodyParser(&numAccount); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		numCard := services.GenerateCardNumber()
		date := services.GenerateDate()
		cvv := services.GenerateCVV()
		if err := h.Storage.CreateCard(context.Background(), numAccount.Number, numCard, date, cvv); err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.Status(fiber.StatusOK).JSON(models.CreateCardResponse{numCard, date, cvv})
	} else {
		return h.HandleRefreshToken(c)
	}
}

func (h *Handler) GetCards(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID != nil {
		var numAccount models.CardRequest
		if err := c.BodyParser(&numAccount); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		response, err := h.Storage.GetCards(context.Background(), numAccount.Number)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.Status(fiber.StatusOK).JSON(response)
	} else {
		return h.HandleRefreshToken(c)
	}
}

func (h *Handler) CreateCredit(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	if userID != nil {
		var request models.CreateCreditRequest
		if err := c.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		err := h.Storage.CreateCredit(context.Background(), request)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.SendStatus(fiber.StatusOK)
	} else {
		return h.HandleRefreshToken(c)
	}
}
