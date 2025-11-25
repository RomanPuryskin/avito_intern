package repository

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type TeamRepository interface {
	/* метод создает в таблице teams новую запись о команде. Принимает на вход
	название команды и возвращает название созданной команды*/
	CreateTeam(ctx context.Context, teamName string) (string, error)

	/* метод возвращает true, если команда уже существует, false, если нет.
	Принимает на вход название команды */
	TeamExists(ctx context.Context, teamName string) (bool, error)
}

type teamPostgresRepository struct {
	Db *pgx.Conn
	sq squirrel.StatementBuilderType
}

func NewTeamPostgresRepository(db *pgx.Conn) *teamPostgresRepository {
	return &teamPostgresRepository{
		Db: db,
		sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (tp *teamPostgresRepository) CreateTeam(ctx context.Context, teamName string) (string, error) {
	// получим транзакцию из контекста
	tx, ok := GetTx(ctx)
	if !ok {
		return "", fmt.Errorf("[TeamRepo | CreateTeam]: can not get pgx.Tx")
	}

	// создадим запись в таблице teams
	query := tp.sq.Insert("teams").
		Columns("team_name").
		Values(teamName)

	sql, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("[TeamRepo | CreateTeam]: %w", err)
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return "", fmt.Errorf("[TeamRepo | CreateTeam]: %w", err)
	}

	return teamName, nil
}

func (tp *teamPostgresRepository) TeamExists(ctx context.Context, teamName string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)"
	var exists bool
	err := tp.Db.QueryRow(ctx, query, teamName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("[TeamRepo | TeamExists]: %w", err)
	}
	return exists, nil
}
