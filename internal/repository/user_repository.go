package repository

import (
	"avito_intern/internal/enteties"
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
	/* метод создает запись о пользователе в табице users. Принимает на вход модель
	enteties.User и возвращает созданного пользователя*/
	CreateUser(ctx context.Context, user *enteties.User) (*enteties.User, error)

	/* метод устанавливает поле status у пользователя. Приинимает user_id и
	новое значение статуса. Возвращает модель нового пользователя enteties.User */
	SetUserStatus(ctx context.Context, userID string, newStatus bool) (*enteties.User, error)

	/* метод возвращает true, если User с заданным user_id существует, false
	если не существует */
	UserExists(ctx context.Context, userID string) (bool, error)

	/* метод возвращает true, если User с заданным username существует, false
	если не существует */
	UserExistsByUsername(ctx context.Context, userName string) (bool, error)

	/* метод возвращает список пользователей, которые находятся в одной команде. Принимает на вход
	название команды, возвращает список моделей enteties.TeamMember*/
	GetTeamMembersByTeamName(ctx context.Context, teamName string) ([]*enteties.TeamMember, error)

	/* метод возвращает название команды, в которой состоит пользователь.
	Принимает на вход user_id*/
	GetUserTeamName(ctx context.Context, userID string) (string, error)
}

type userPostgresRepository struct {
	Db *pgx.Conn
	sq squirrel.StatementBuilderType
}

func NewUserPostgresRepository(db *pgx.Conn) *userPostgresRepository {
	return &userPostgresRepository{
		Db: db,
		sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (urp *userPostgresRepository) CreateUser(ctx context.Context, user *enteties.User) (*enteties.User, error) {

	// получим транзакцию из контекста
	tx, ok := GetTx(ctx)
	if !ok {
		return nil, fmt.Errorf("[UserRepo | CreateUser]: can not get pgx.Tx")
	}

	query := urp.sq.Insert("users").
		Columns("user_id", "username", "team_name", "is_active").
		Values(user.UserID, user.UserName, user.TeamName, user.IsActive)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("[UserRepo | CreateUser]: %w", err)
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("[UserRepo | CreateUser]: %w", err)
	}

	return user, nil
}

func (urp *userPostgresRepository) SetUserStatus(ctx context.Context, userID string, newStatus bool) (*enteties.User, error) {

	query := urp.sq.Update("users").
		Set("is_active", newStatus).
		Where(squirrel.Eq{"user_id": userID}).
		Suffix("RETURNING username , team_name,is_active")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("[UserRepo | SetUserStatus]: %w", err)
	}

	row := urp.Db.QueryRow(ctx, sql, args...)

	var userResponce enteties.User
	userResponce.UserID = userID

	err = row.Scan(&userResponce.UserName, &userResponce.TeamName, &userResponce.IsActive)
	if err != nil {
		return nil, fmt.Errorf("[UserRepo | SetUserStatus]: %w", err)
	}

	return &userResponce, nil

}

func (urp *userPostgresRepository) UserExists(ctx context.Context, userID string) (bool, error) {

	query := "SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)"
	var exists bool
	err := urp.Db.QueryRow(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("[UserRepo | UserExists]: %w", err)
	}
	return exists, nil
}

func (urp *userPostgresRepository) UserExistsByUsername(ctx context.Context, userName string) (bool, error) {

	query := "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)"
	var exists bool
	err := urp.Db.QueryRow(ctx, query, userName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("[UserRepo | UserExistsByUsername]: %w", err)
	}
	return exists, nil
}

func (urp *userPostgresRepository) GetTeamMembersByTeamName(ctx context.Context, teamName string) ([]*enteties.TeamMember, error) {

	// получим транзакцию из контекста
	tx, ok := GetTx(ctx)
	if !ok {
		return nil, fmt.Errorf("[UserRepo | GetTeamMembersByTeamName]: can not get pgx.Tx")
	}

	teamMembers := make([]*enteties.TeamMember, 0)

	query := urp.sq.Select(
		"user_id",
		"username",
		"is_active").
		From("users").
		Where(squirrel.Eq{"team_name": teamName})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("[UserRepo | GetTeamMembersByTeamName]: %w", err)
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("[UserRepo | GetTeamMembersByTeamName]: %w", err)
	}

	for rows.Next() {
		var tm enteties.TeamMember

		rows.Scan(&tm.UserID, &tm.UserName, &tm.IsActive)
		if err != nil {
			return nil, fmt.Errorf("[UserRepo | GetTeamMembersByTeamName]: %w", err)
		}

		teamMembers = append(teamMembers, &tm)
	}

	return teamMembers, nil
}

func (urp *userPostgresRepository) GetUserTeamName(ctx context.Context, userID string) (string, error) {
	var teamName string
	query := `SELECT team_name FROM users WHERE user_id = $1`

	err := urp.Db.QueryRow(ctx, query, userID).Scan(&teamName)
	if err != nil {
		return "", fmt.Errorf("[UserRepo | GetUsersTeamName]: %w", err)
	}

	return teamName, nil
}
