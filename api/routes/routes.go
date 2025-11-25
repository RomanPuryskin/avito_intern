package routes

import (
	"avito_intern/api/handlers"

	"github.com/gofiber/fiber/v2"
)

func InitUserRoutes(app *fiber.App, h *handlers.UserHandler) {
	api := app.Group("/users")
	api.Post("/setIsActive", h.SetIsActive)
	api.Get("/getReview", h.GetReview)
}

func InitTeamRoutes(app *fiber.App, h *handlers.TeamHandler) {
	api := app.Group("team")
	api.Post("/add", h.CreateTeam)
	api.Get("/get", h.GetTeam)
}

func InitPRRoutes(app *fiber.App, h *handlers.PRHandler) {
	api := app.Group("/pullRequest")
	api.Post("/create", h.CreatePR)
	api.Post("/merge", h.MergePR)
	api.Post("/reassign", h.ReassignPR)
}
