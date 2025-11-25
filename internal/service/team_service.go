package service

import (
	"avito_intern/internal/enteties"
	"avito_intern/internal/repository"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var (
	ErrorTeamExists   = errors.New("team already exists")
	ErrorTeamNotFound = errors.New("team not found")
)

//go:generate mockgen -source=team_service.go -destination=../../mocks/team_service.go -package=mocks
type TeamService interface {
	/* метод создает запись о новой команде в таблице teams, а так же создает записи
	о пользователях, которые в ней состоят в таблице users. Принимает на вход модель
	enteties.Team и возвращает модель созданной команды enteties.Team*/
	CreateTeam(ctx context.Context, team *enteties.Team) (*enteties.Team, error)

	/* метод возвращает инфо о команде в виде модели enteties.Team.
	Принимает на вход название команды */
	GetTeam(ctx context.Context, teamName string) (*enteties.Team, error)
}

type teamService struct {
	Db       *pgx.Conn
	UserRepo repository.UserRepository
	TeamRepo repository.TeamRepository
}

func NewTeamService(db *pgx.Conn, userRepo repository.UserRepository, teamRepo repository.TeamRepository) *teamService {
	return &teamService{
		Db:       db,
		UserRepo: userRepo,
		TeamRepo: teamRepo,
	}
}

func (ts *teamService) CreateTeam(ctx context.Context, team *enteties.Team) (*enteties.Team, error) {

	// проверим, существует ли уже команда с таким именем
	exists, err := ts.TeamRepo.TeamExists(ctx, team.TeamName)
	if err != nil {
		return nil, fmt.Errorf("[TeamService| CreateTeam]: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("[TeamService| CreateTeam]: %w", ErrorTeamExists)
	}

	tx, err := ts.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("[TeamService| CreateTeam]: %w", err)
	}
	defer tx.Rollback(ctx)

	// добавим tx в контекст
	ctx = repository.WithTx(ctx, tx)

	// созадим запись в таблице team
	_, err = ts.TeamRepo.CreateTeam(ctx, team.TeamName)
	if err != nil {
		return nil, fmt.Errorf("[TeamService| CreateTeam]: %w", err)
	}

	for _, tm := range team.Members {
		user := &enteties.User{
			UserID:   tm.UserID,
			UserName: tm.UserName,
			TeamName: team.TeamName,
			IsActive: tm.IsActive,
		}

		// проверим, существует ли пользователь с таким id
		exists, err := ts.UserRepo.UserExists(ctx, user.UserID)
		if err != nil {
			return nil, fmt.Errorf("[TeamService| CreateTeam]: %w", err)
		}

		if exists {
			return nil, fmt.Errorf("[TeamService| CreateTeam]: %w", ErrorUserAlreadyExists)
		}

		// проверим, существует ли пользователь с таким username
		exists, err = ts.UserRepo.UserExistsByUsername(ctx, user.UserName)
		if err != nil {
			return nil, fmt.Errorf("[TeamService| CreateTeam]: %w", err)
		}

		if exists {
			return nil, fmt.Errorf("[TeamService| CreateTeam]: %w", ErrorUserAlreadyExistsByUserName)
		}

		// вставим пользователя в таблицу
		_, err = ts.UserRepo.CreateUser(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("[TeamService| CreateTeam]: %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("[TeamService| CreateTeam]: %w", err)
	}

	return team, nil
}

func (ts *teamService) GetTeam(ctx context.Context, teamName string) (*enteties.Team, error) {

	// проверим существование команды
	exists, err := ts.TeamRepo.TeamExists(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("[TeamService | GetTeam]: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("[TeamService | GetTeam]: %w", ErrorTeamNotFound)
	}

	tx, err := ts.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("[TeamService | GetTeam]: %w", err)
	}
	defer tx.Rollback(ctx)

	// добавим tx в контекст
	ctx = repository.WithTx(ctx, tx)

	// получим всех teamMembers по teamName
	teamMembersPtrs, err := ts.UserRepo.GetTeamMembersByTeamName(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("[TeamService | GetTeam]: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("[TeamService | GetTeam]: %w", err)
	}

	teamMembers := make([]enteties.TeamMember, len(teamMembersPtrs))
	for ind, tm := range teamMembersPtrs {
		teamMembers[ind] = *tm
	}

	return &enteties.Team{
		TeamName: teamName,
		Members:  teamMembers,
	}, nil
}
