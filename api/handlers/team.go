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

type TeamHandler struct {
	Logger  *slog.Logger
	Service service.TeamService
}

func NewTeamHandler(log *slog.Logger, service service.TeamService) *TeamHandler {
	return &TeamHandler{
		Logger:  log,
		Service: service,
	}
}

func (th *TeamHandler) CreateTeam(c *fiber.Ctx) error {

	var team enteties.Team

	// парсинг json request
	err := c.BodyParser(&team)
	if err != nil {
		th.Logger.Error("failed parse team", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(errs.ErrorInvaidInputFormat)
	}

	// валидация полученной структуры
	err = utils.ValidateStruct(&team)
	if err != nil {
		th.Logger.Error("failed validate team", "error", err, "request", team)
		return c.Status(fiber.StatusBadRequest).JSON(errs.ErrorInvaidInput)
	}

	// провалидируем каждого пользователя
	for _, tm := range team.Members {
		err = utils.ValidateStruct(&tm)
		if err != nil {
			th.Logger.Error("failed validate team", "error", err, "request", team)
			return c.Status(fiber.StatusBadRequest).JSON(errs.ErrorInvaidInput)
		}
	}

	// используем контекст от fiber для всех операций (он уже правильно настроен)
	ctx := c.Context()

	respTeam, err := th.Service.CreateTeam(ctx, &team)
	if err != nil {
		slog.Error("failed create team", "error", err, "input", team)
		switch {
		case errors.Is(err, service.ErrorUserAlreadyExists):
			return c.Status(fiber.StatusBadRequest).JSON(errs.ErrorUserAlreadyExists)
		case errors.Is(err, service.ErrorUserAlreadyExistsByUserName):
			return c.Status(fiber.StatusBadRequest).JSON(errs.ErrorUserAlreadyExistsByUserName)
		case errors.Is(err, service.ErrorTeamExists):
			return c.Status(fiber.StatusBadRequest).JSON(errs.ErrorTeamAlreadyExists)
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(errs.ErrorInternal)
		}
	}

	slog.Info("success team created", "input", team, "responce", respTeam)
	return c.Status(fiber.StatusCreated).JSON(respTeam)
}

func (th *TeamHandler) GetTeam(c *fiber.Ctx) error {

	teamName := c.Query("team_name", "")
	if teamName == "" {
		slog.Error("failed get team", "query", teamName)
		return c.Status(fiber.StatusBadRequest).JSON(errs.ErrorInvaidInput)
	}

	// используем контекст от fiber для всех операций (он уже правильно настроен)
	ctx := c.Context()

	team, err := th.Service.GetTeam(ctx, teamName)
	if err != nil {
		slog.Error("failed get team", "error", err, "input", teamName)
		switch {
		case errors.Is(err, service.ErrorTeamNotFound):
			return c.Status(fiber.StatusNotFound).JSON(errs.ErrorTeamNotFound)
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(errs.ErrorInternal)
		}
	}

	slog.Info("Success got team", "input", teamName, "responce", team)
	return c.Status(fiber.StatusOK).JSON(team)
}
