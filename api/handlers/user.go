package handlers

import (
	"avito_intern/api/errs"
	"avito_intern/internal/enteties"
	"avito_intern/internal/service"
	"avito_intern/internal/utils"
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	Logger  *slog.Logger
	Service service.UserService
}

func NewUserHandler(log *slog.Logger, service service.UserService) *UserHandler {
	return &UserHandler{
		Logger:  log,
		Service: service,
	}
}

func (uh *UserHandler) SetIsActive(c *fiber.Ctx) error {
	var request enteties.RequestUserToSetActive

	// парсинг json request
	err := c.BodyParser(&request)
	if err != nil {
		uh.Logger.Error("failed parse request user to set status", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(errs.ErrorInvaidInputFormat)
	}

	// валидация полученной структуры
	err = utils.ValidateStruct(&request)
	if err != nil {
		uh.Logger.Error("failed validate request user to set status", "error", err, "request", request)
		return c.Status(fiber.StatusBadRequest).JSON(errs.ErrorInvaidInput)
	}

	// используем контекст от fiber для всех операций (он уже правильно настроен)
	ctx := c.Context()

	userResp, err := uh.Service.SetIsActive(ctx, request.UserID, request.IsActive)
	if err != nil {

		slog.Error("failed set status", "error", err, "input", request)
		switch {
		case errors.Is(err, service.ErrorUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(errs.ErrorUserNotFound)

		default:
			return c.Status(fiber.StatusInternalServerError).JSON(errs.ErrorInternal)
		}
	}

	slog.Info("success status set", "input", request, "responce", userResp)
	return c.Status(fiber.StatusOK).JSON(userResp)

}

func (uh *UserHandler) GetReview(c *fiber.Ctx) error {

	userID := c.Query("user_id", "")
	if userID == "" {
		slog.Error("failed get user", "query", userID)
		return c.Status(fiber.StatusBadRequest).JSON(errs.ErrorInvaidInput)
	}

	// используем контекст от fiber для всех операций (он уже правильно настроен)
	ctx := c.Context()

	userReviews, err := uh.Service.GetReviews(ctx, userID)
	if err != nil {

		slog.Error("failed get reviews", "error", err, "input", userID)
		switch {
		case errors.Is(err, service.ErrorUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(errs.ErrorUserNotFound)

		default:
			return c.Status(fiber.StatusInternalServerError).JSON(errs.ErrorInternal)
		}
	}

	slog.Info("success got reviews", "input", userID, "responce", userReviews)
	return c.Status(fiber.StatusOK).JSON(userReviews)
}
