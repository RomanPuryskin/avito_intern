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
	ErrorUserNotFound                = errors.New("user not found")
	ErrorUserAlreadyExists           = errors.New("user already exists")
	ErrorUserAlreadyExistsByUserName = errors.New("user with username already exists")
)

//go:generate mockgen -source=user_service.go -destination=../../mocks/user_service.go -package=mocks
type UserService interface {
	/* метод устанавливает поле is_active у пользователя. Принимает на вход user_id,
	значение is_active, которое нужно установить. Возвращает модель enteties.User*/
	SetIsActive(ctx context.Context, userID string, status bool) (*enteties.User, error)

	/* метод возвращает pull request' ы, где пользователь назачен ревьюером в формате
	модели enteties.UserReviewers. Принимает на вход user_id*/
	GetReviews(ctx context.Context, userID string) (*enteties.UserReviews, error)
}

type userService struct {
	Db       *pgx.Conn
	UserRepo repository.UserRepository
	PRRepo   repository.PRRepository
}

func NewUserService(db *pgx.Conn, userRepo repository.UserRepository, prRepo repository.PRRepository) *userService {
	return &userService{
		Db:       db,
		UserRepo: userRepo,
		PRRepo:   prRepo,
	}
}

func (us *userService) SetIsActive(ctx context.Context, userID string, status bool) (*enteties.User, error) {

	// проверяем существование пользователя
	exists, err := us.UserRepo.UserExists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("[UserService | setIsActive]: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("[UserService | setIsActive]: %w", ErrorUserNotFound)
	}

	// если пользователь есть, меняем статус
	userResp, err := us.UserRepo.SetUserStatus(ctx, userID, status)
	if err != nil {
		return nil, fmt.Errorf("[UserService | setIsActive]: %w", err)
	}

	return userResp, nil
}

func (us *userService) GetReviews(ctx context.Context, userID string) (*enteties.UserReviews, error) {
	// проверяем существование пользователя
	exists, err := us.UserRepo.UserExists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("[UserService | GetReviews]: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("[UserService | GetReviews]: %w", ErrorUserNotFound)
	}

	tx, err := us.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("[UserService | GetReviews]: %w", err)
	}
	defer tx.Rollback(ctx)

	// добавим tx в контекст
	ctx = repository.WithTx(ctx, tx)

	shortPRptrs, err := us.PRRepo.GetAllPRByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("[UserService | GetReviews]: %w", err)

	}

	shortPRs := make([]enteties.PullRequestShort, len(shortPRptrs))

	for ind, pr := range shortPRptrs {
		shortPRs[ind] = *pr
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("[UserService | GetReviews]: %w", err)
	}

	return &enteties.UserReviews{
		UserID:       userID,
		PullRequests: shortPRs,
	}, nil
}
