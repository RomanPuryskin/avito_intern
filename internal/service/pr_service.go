package service

import (
	"avito_intern/internal/enteties"
	"avito_intern/internal/repository"
	"context"
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/jackc/pgx/v5"
)

var (
	ErrorPRAlreadyExists       = errors.New("PR already exists")
	ErrorPRNotFound            = errors.New("PR not found")
	ErrorPRIsMerged            = errors.New("PR is merged")
	ErrorUserNotAssigned       = errors.New("user not assigned to PR")
	ErrorNoCandidateToReassign = errors.New("no candiate to reassign")
)

//go:generate mockgen -source=pr_service.go -destination=../../mocks/pr_service.go -package=mocks
type PRService interface {
	/* метод создает новый pull request, занося информацию в таблицу pull_requests
	и автоматически определяя на него до двух ревьюеров из команды автора.
	Принимает на вход модель enteties.CreatePullRequest, возвращает инфо о созданном
	pull request в виде модели enteties.PullRequest */
	CreatePR(ctx context.Context, pr *enteties.CreatePullRequest) (*enteties.PullRequest, error)

	/* идемпотентный метод cтавит статус pull_request в MERGED. Принимаем на вход запрос
	в виде модели enteties.MergePullRequest , возвращает pull request enteties.PullRequest*/
	MergePR(ctx context.Context, mergeReq *enteties.MergePullRequest) (*enteties.PullRequest, error)

	/* метод переназначает ревьюера на pull_request. Принимает на вход модель
	enteties.ReassignPullRequest, возвращает модель enteties.ReassignPullRequestResponce*/
	ReassignPR(ctx context.Context, resp *enteties.ReassignPullRequest) (*enteties.ReassignPullRequestResponce, error)
}

type prService struct {
	Db       *pgx.Conn
	UserRepo repository.UserRepository
	TeamRepo repository.TeamRepository
	PRRepo   repository.PRRepository
}

func NewPRService(db *pgx.Conn, userRepo repository.UserRepository, teamRepo repository.TeamRepository,
	prRepo repository.PRRepository) *prService {
	return &prService{
		Db:       db,
		UserRepo: userRepo,
		TeamRepo: teamRepo,
		PRRepo:   prRepo,
	}
}

func (prs *prService) CreatePR(ctx context.Context, pr *enteties.CreatePullRequest) (*enteties.PullRequest, error) {

	// проверим, существует ли уже PR с таким id
	exists, err := prs.PRRepo.PRExists(ctx, pr.PullRequestID)
	if err != nil {
		return nil, fmt.Errorf("[PRService | CreatePR]: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("[PRService | CreatePR]: %w", ErrorPRAlreadyExists)
	}
	// проверим что пользователь с таким id есть
	exists, err = prs.UserRepo.UserExists(ctx, pr.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("[PRService | CreatePR]: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("[PRService | CreatePR]: %w", ErrorUserNotFound)
	}

	// найдем команду автора
	teamName, err := prs.UserRepo.GetUserTeamName(ctx, pr.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("[PRService | CreatePR]: %w", err)
	}

	tx, err := prs.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("[PRService | CreatePR]: %w", err)
	}
	defer tx.Rollback(ctx)

	// добавим tx в контекст
	ctx = repository.WithTx(ctx, tx)

	// занесем инфо в таблицу
	prShort, err := prs.PRRepo.CreatePR(ctx, pr)
	if err != nil {
		return nil, fmt.Errorf("[PRService | CreatePR]: %w", err)
	}

	// назначим ревьюеров ( получим список юзеров которые:
	//1) в той же команде 2) со статусом is_active)

	// найдем сначала всех сокомандников
	teamMembers, err := prs.UserRepo.GetTeamMembersByTeamName(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("[PRService | CreatePR]: %w", err)
	}

	// получим список доступных к назначению на ревьюера (статус is_active и исполючим автора)
	futureReviewers := []string{}
	for _, tm := range teamMembers {
		if tm.IsActive == true && tm.UserID != pr.AuthorID {
			futureReviewers = append(futureReviewers, tm.UserID)
		}
	}

	// назначим ревьюеров
	reviewers := []string{}

	// если их меньше двух доступных, назначаем их ( 0 или 1 )
	if len(futureReviewers) < 2 {
		reviewers = futureReviewers
	} else {
		// если ревьюеров >= 2 случайным образом выберем количество ревьюеров (0 , 1 , 2)
		randomNumber := rand.IntN(3)

		for i := 0; i < randomNumber; i++ {
			// выберем случайного пользователя из списка доступных
			randomIndexRev := rand.Int32N(int32(len(futureReviewers)))
			reviewers = append(reviewers, futureReviewers[randomIndexRev])

			// уберем из списка доступных, чтобы больше его не выбрать
			futureReviewers = append(futureReviewers[:randomIndexRev], futureReviewers[randomIndexRev+1:]...)
		}
	}

	// занесем назначенных ревьюеров
	err = prs.PRRepo.SetReviewersBatch(ctx, pr.PullRequestID, reviewers)
	if err != nil {
		return nil, fmt.Errorf("[PRService | CreatePR]: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("[PRService | CreatePR]: %w", err)
	}

	return &enteties.PullRequest{
		PullRequestID:     prShort.PullRequestID,
		PulRequestName:    prShort.PulRequestName,
		AuthorID:          prShort.AuthorID,
		Status:            prShort.Status,
		AssignedReviewers: reviewers,
	}, nil

}

func (prs *prService) MergePR(ctx context.Context, mergeReq *enteties.MergePullRequest) (*enteties.PullRequest, error) {
	// проверим, существует ли уже PR с таким id
	exists, err := prs.PRRepo.PRExists(ctx, mergeReq.PullRequestID)
	if err != nil {
		return nil, fmt.Errorf("[PRService | MergePR]: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("[PRService | MergePR]: %w", ErrorPRNotFound)
	}

	tx, err := prs.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("[PRService | MergePR]: %w", err)
	}
	defer tx.Rollback(ctx)

	// добавим tx в контекст
	ctx = repository.WithTx(ctx, tx)

	// мерджим
	respPR, err := prs.PRRepo.MergePR(ctx, mergeReq.PullRequestID)
	if err != nil {
		return nil, fmt.Errorf("[PRService | MergePR]: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("[PRService | MergePR]: %w", err)
	}

	return respPR, nil
}

func (prs *prService) ReassignPR(ctx context.Context, resp *enteties.ReassignPullRequest) (*enteties.ReassignPullRequestResponce, error) {
	// проверим, существует ли pr
	exists, err := prs.PRRepo.PRExists(ctx, resp.PullRequestID)
	if err != nil {
		return nil, fmt.Errorf("[PRService | ReassignPR]: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("[PRService | ReassignPR]: %w", ErrorPRNotFound)
	}

	// проверим, существует ли пользователь
	exists, err = prs.UserRepo.UserExists(ctx, resp.OldUserID)
	if err != nil {
		return nil, fmt.Errorf("[PRService | ReassignPR]: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("[PRService | ReassignPR]: %w", ErrorUserNotFound)
	}

	// проверим, не смерджен ли уже pr
	isMerged, err := prs.PRRepo.IsMerged(ctx, resp.PullRequestID)
	if err != nil {
		return nil, fmt.Errorf("[PRService | ReassignPR]: %w", err)
	}
	if isMerged {
		return nil, fmt.Errorf("[PRService | ReassignPR]: %w", ErrorPRIsMerged)
	}

	// проверим, назначен ли пользователь ревьюером на этот pr
	isReviewed, err := prs.PRRepo.IsUserAssignedToPR(ctx, resp.OldUserID, resp.PullRequestID)
	if err != nil {
		return nil, fmt.Errorf("[PRService | ReassignPR]: %w", err)
	}
	if !isReviewed {
		return nil, fmt.Errorf("[PRService | ReassignPR]: %w", ErrorUserNotAssigned)
	}

	// найдем команду заменяемого
	teamName, err := prs.UserRepo.GetUserTeamName(ctx, resp.OldUserID)
	if err != nil {
		return nil, fmt.Errorf("[PRService | ReassignPR]: %w", err)
	}

	// найдем автора pr, чтобы его не поставить ревьюером
	author, err := prs.PRRepo.GetAuthorPR(ctx, resp.PullRequestID)
	if err != nil {
		return nil, fmt.Errorf("[PRService | ReassignPR]: %w", err)
	}

	tx, err := prs.Db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("[PRService | ReassignPR]: %w", err)
	}
	defer tx.Rollback(ctx)

	// добавим tx в контекст
	ctx = repository.WithTx(ctx, tx)

	// найдем сначала всех сокомандников
	teamMembers, err := prs.UserRepo.GetTeamMembersByTeamName(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("[PRService | ReassignPR]: %w", err)
	}

	// получим список доступных к назначению на ревьюера (статус is_active и исполючим автора)
	futureReviewers := []string{}
	for _, tm := range teamMembers {
		if tm.IsActive == true && tm.UserID != resp.OldUserID && tm.UserID != author {
			futureReviewers = append(futureReviewers, tm.UserID)
		}
	}

	// не доступных сокомандников для замены
	if len(futureReviewers) == 0 {
		return nil, fmt.Errorf("[PRService | ReassignPR]: %w", ErrorNoCandidateToReassign)
	}

	// выберем случайного пользователя из списка доступных
	randomIndexRev := rand.Int32N(int32(len(futureReviewers)))

	newReviewer := futureReviewers[randomIndexRev]

	// перезапишем связь в таблице assigned_reviewers
	err = prs.PRRepo.ReassignReviewer(ctx, resp.PullRequestID, resp.OldUserID, newReviewer)
	if err != nil {
		return nil, fmt.Errorf("[PRService | ReassignPR]: %w", err)
	}

	// получим новый PR
	pr, err := prs.PRRepo.GetPR(ctx, resp.PullRequestID)
	if err != nil {
		return nil, fmt.Errorf("[PRService | ReassignPR]: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("[PRService | ReassignPR]: %w", err)
	}

	return &enteties.ReassignPullRequestResponce{
		PR:         *pr,
		ReplacedBy: newReviewer,
	}, nil
}
