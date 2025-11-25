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

type PRHandler struct {
	Logger  *slog.Logger
	Service service.PRService
}

func NewPRHandler(log *slog.Logger, service service.PRService) *PRHandler {
	return &PRHandler{
		Logger:  log,
		Service: service,
	}
}

func (prh *PRHandler) CreatePR(c *fiber.Ctx) error {

	var pr enteties.CreatePullRequest

	// парсинг json request
	err := c.BodyParser(&pr)
	if err != nil {
		prh.Logger.Error("failed parse pr to create", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(errs.ErrorInvaidInputFormat)
	}

	// валидация полученной структуры
	err = utils.ValidateStruct(&pr)
	if err != nil {
		prh.Logger.Error("failed validate pr to create", "error", err, "request", pr)
		return c.Status(fiber.StatusBadRequest).JSON(errs.ErrorInvaidInput)
	}

	// используем контекст от fiber для всех операций (он уже правильно настроен)
	ctx := c.Context()

	respPR, err := prh.Service.CreatePR(ctx, &pr)
	if err != nil {
		// обработка ошибок
		slog.Error("failed create pr", "error", err, "input", pr)
		switch {
		case errors.Is(err, service.ErrorUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(errs.ErrorUserNotFound)
		case errors.Is(err, service.ErrorPRAlreadyExists):
			return c.Status(fiber.StatusConflict).JSON(errs.ErrorPRAlreadyExists)
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(errs.ErrorInternal)
		}

	}

	slog.Info("succesc PR created", "input", pr, "responce", respPR)
	return c.Status(fiber.StatusCreated).JSON(respPR)
}

func (prh *PRHandler) MergePR(c *fiber.Ctx) error {

	var mergePR enteties.MergePullRequest

	// парсинг json request
	err := c.BodyParser(&mergePR)
	if err != nil {
		prh.Logger.Error("failed parse pr to merge", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(errs.ErrorInvaidInputFormat)
	}

	// валидация полученной структуры
	err = utils.ValidateStruct(&mergePR)
	if err != nil {
		prh.Logger.Error("failed validate pr to merge", "error", err, "request", mergePR)
		return c.Status(fiber.StatusBadRequest).JSON(errs.ErrorInvaidInput)
	}

	// используем контекст от fiber для всех операций (он уже правильно настроен)
	ctx := c.Context()

	respPR, err := prh.Service.MergePR(ctx, &mergePR)
	if err != nil {
		// обработка ошибок
		slog.Error("failed merge pr", "error", err, "input", mergePR)
		switch {
		case errors.Is(err, service.ErrorPRNotFound):
			return c.Status(fiber.StatusNotFound).JSON(errs.ErrorPRNotFound)
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(errs.ErrorInternal)
		}
	}

	slog.Info("success PR merged", "input", mergePR, "responce", respPR)
	return c.Status(fiber.StatusOK).JSON(respPR)
}

func (prh *PRHandler) ReassignPR(c *fiber.Ctx) error {

	var reassignPR enteties.ReassignPullRequest

	// парсинг json request
	err := c.BodyParser(&reassignPR)
	if err != nil {
		prh.Logger.Error("failed parse pr to reassign", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(errs.ErrorInvaidInputFormat)
	}

	// валидация полученной структуры
	err = utils.ValidateStruct(&reassignPR)
	if err != nil {
		prh.Logger.Error("failed validate pr to reassign", "error", err, "request", reassignPR)
		return c.Status(fiber.StatusBadRequest).JSON(errs.ErrorInvaidInput)
	}

	// используем контекст от fiber для всех операций (он уже правильно настроен)
	ctx := c.Context()

	resp, err := prh.Service.ReassignPR(ctx, &reassignPR)
	if err != nil {
		slog.Error("failed reassign pr", "error", err, "input", reassignPR)
		switch {
		case errors.Is(err, service.ErrorPRNotFound):
			return c.Status(fiber.StatusNotFound).JSON(errs.ErrorPRNotFound)
		case errors.Is(err, service.ErrorUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(errs.ErrorUserNotFound)
		case errors.Is(err, service.ErrorPRIsMerged):
			return c.Status(fiber.StatusConflict).JSON(errs.ErrorPRMerged)
		case errors.Is(err, service.ErrorUserNotAssigned):
			return c.Status(fiber.StatusConflict).JSON(errs.ErrorUserNotAssigned)
		case errors.Is(err, service.ErrorNoCandidateToReassign):
			return c.Status(fiber.StatusConflict).JSON(errs.ErrorNoCandidateToReassign)
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(errs.ErrorInternal)
		}
	}

	slog.Info("success PR reassigned", "input", reassignPR, "responce", resp)
	return c.Status(fiber.StatusOK).JSON(resp)
}
