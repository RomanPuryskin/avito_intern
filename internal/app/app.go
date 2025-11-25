package app

import (
	"avito_intern/api/handlers"
	"avito_intern/api/routes"
	"avito_intern/internal/config"
	"avito_intern/internal/database/postgres"
	"avito_intern/internal/repository"
	"avito_intern/internal/service"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type App struct {
	Cfg      *config.Config
	FiberApp *fiber.App
	Storage  *pgx.Conn
	Logger   *slog.Logger
}

func InitNewApp(ctx context.Context, cfg *config.Config, log *slog.Logger) *App {

	// подключаемся к DB
	conn, err := postgres.NewPostgresDB(ctx, cfg)
	if err != nil {
		slog.Error("Failed to connect postgres DB", "error", err)
		os.Exit(1)
	}

	log.Info("Successfully connected to postgres DB")

	// запускаем миграции
	err = postgres.RunMigrations(cfg)
	if err != nil {
		log.Error("Failed run migrations",
			"error", err)
		os.Exit(1)
	}
	log.Info("Successfully ran migrations")

	// создание репозиториев
	userRepo := repository.NewUserPostgresRepository(conn)
	teamRepo := repository.NewTeamPostgresRepository(conn)
	prRepo := repository.NewPRPostgresRepository(conn)

	// создание сервисов
	userService := service.NewUserService(conn, userRepo, prRepo)
	teamService := service.NewTeamService(conn, userRepo, teamRepo)
	prService := service.NewPRService(conn, userRepo, teamRepo, prRepo)

	// создание приложения fiber
	app := fiber.New(fiber.Config{
		Prefork: false,
	})

	// подключение хэндлеров
	userHanlder := handlers.NewUserHandler(log, userService)
	teamHandler := handlers.NewTeamHandler(log, teamService)
	prHandler := handlers.NewPRHandler(log, prService)

	// подключение роутов
	routes.InitUserRoutes(app, userHanlder)
	routes.InitTeamRoutes(app, teamHandler)
	routes.InitPRRoutes(app, prHandler)

	return &App{
		Cfg:      cfg,
		FiberApp: app,
		Storage:  conn,
		Logger:   log,
	}
}

func (a *App) Start(ctx context.Context) {
	a.Logger.Info("App started", "port", a.Cfg.Server.ServerPort)

	go func() {
		err := a.FiberApp.Listen(fmt.Sprintf(":%s", a.Cfg.Server.ServerPort))
		if err != nil {
			a.Logger.Error("Failed to start app",
				"error", err)
			os.Exit(1)
		}
	}()
}

func (a *App) Stop(ctx context.Context) error {
	a.Logger.Info("[!] Shutting down...")

	var stopErr = errors.New("")

	// закрываем соединение БД
	if err := postgres.ClosePostgresDB(ctx, a.Storage); err != nil {
		errors.Join(stopErr, err)
	}

	// закрываем соединение с сервером
	if err := a.FiberApp.ShutdownWithContext(ctx); err != nil {
		errors.Join(stopErr, err)
	}

	if stopErr.Error() != "" {
		return stopErr
	}
	return nil
}
