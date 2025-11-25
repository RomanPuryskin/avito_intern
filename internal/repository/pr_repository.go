package repository

import (
	"avito_intern/internal/enteties"
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type PRRepository interface {
	/* метод создает запись о pull request в таблице pull_requests. Принимает на вход
	модель enteties.CreatePullRequest , возвращает модель enteties.PullRequestShort*/
	CreatePR(ctx context.Context, pr *enteties.CreatePullRequest) (*enteties.PullRequestShort, error)

	/* метод возвращает true, если pull request с заданным id существует, false иначе.
	Принимает на вход pull_request_id*/
	PRExists(ctx context.Context, id string) (bool, error)

	/* метод заносит в таблицу assigned_reviewers записи сразу пачкой.
	Принимает на вход pull_request_id и список user_id ревьюеров*/
	SetReviewersBatch(ctx context.Context, PR_id string, usersID []string) error

	/* метод возвращает true, если у pull_request status MERGED, иначе false.
	Принимает на вход pull_request_id*/
	IsMerged(ctx context.Context, PR_id string) (bool, error)

	/* идемпотентный метод, который устанавливает status у pull_request в значение MERGED.
	Принимает на вход pull_request_id, возвращает модель enteties.PullRequest*/
	MergePR(ctx context.Context, PR_id string) (*enteties.PullRequest, error)

	/* метод возвращает информацию о pull_request в виде модели enteties.PullRequest.
	Принимает на вход pull_request_id*/
	GetPR(ctx context.Context, PR_id string) (*enteties.PullRequest, error)

	/* метод возвращает все pull_request, на которые пользователь назначен ревьюером.
	Принимает на вход user_id, возвращает список моделей enteties.PullRequestShort */
	GetAllPRByUserID(ctx context.Context, user_id string) ([]*enteties.PullRequestShort, error)

	/* метод возвращает true, если пользователь назначен ревьюером на pull request, иначе
	false. Принимает на вход user_id и pull_request_id*/
	IsUserAssignedToPR(ctx context.Context, userID, prID string) (bool, error)

	/* метод заменяет в таблице assigned_reviewers ревьюера на pull request. Принимает
	на вход pull_request, user_id заменяемого пользователя и user_id замещающего*/
	ReassignReviewer(ctx context.Context, prID, oldUserID, newUserID string) error

	/* метод возвращает автора pull request. Принимает на вход pull_request_id*/
	GetAuthorPR(ctx context.Context, prID string) (string, error)
}

type prPostgresRepository struct {
	Db *pgx.Conn
	sq squirrel.StatementBuilderType
}

func NewPRPostgresRepository(db *pgx.Conn) *prPostgresRepository {
	return &prPostgresRepository{
		Db: db,
		sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (prp *prPostgresRepository) CreatePR(ctx context.Context, pr *enteties.CreatePullRequest) (*enteties.PullRequestShort, error) {

	prShort := enteties.PullRequestShort{
		PullRequestID:  pr.PullRequestID,
		PulRequestName: pr.PullRequestName,
		AuthorID:       pr.AuthorID,
	}

	// получим транзакцию из контекста
	tx, ok := GetTx(ctx)
	if !ok {
		return nil, fmt.Errorf("[PRRepo | CreatePR]: can not get pgx.Tx")
	}

	query := prp.sq.Insert("pull_requests").
		Columns("pull_request_id", "pull_request_name", "author_id", "status").
		Values(pr.PullRequestID, pr.PullRequestName, pr.AuthorID, enteties.PullRequestStatusOpen).
		Suffix("RETURNING status")

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("[PRRepo | CreatePR]: %w", err)
	}

	row := tx.QueryRow(ctx, sql, args...)

	err = row.Scan(&prShort.Status)
	if err != nil {
		return nil, fmt.Errorf("[PRRepo | CreatePR]: %w", err)
	}

	return &prShort, nil
}

func (prp *prPostgresRepository) PRExists(ctx context.Context, id string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)"
	var exists bool
	err := prp.Db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("[PRRepo | PRExists]: %w", err)
	}
	return exists, nil
}

func (prp *prPostgresRepository) SetReviewersBatch(ctx context.Context, PR_id string, usersID []string) error {

	batch := &pgx.Batch{}

	// получим транзакцию из контекста
	tx, ok := GetTx(ctx)
	if !ok {
		return fmt.Errorf("[PRRepo | SetReviewersBatch]: can not get pgx.Tx")
	}

	for _, user := range usersID {
		batch.Queue(`INSERT INTO assigned_reviewers(pull_request_id, user_id)
		VALUES ($1, $2)`, PR_id, user)
	}
	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("[PRRepo | SetReviewersBatch]: %w", err)
		}
	}

	return nil
}

func (prp *prPostgresRepository) IsMerged(ctx context.Context, PR_id string) (bool, error) {
	var status string

	query := prp.sq.Select("status").
		From("pull_requests").
		Where(squirrel.Eq{"pull_request_id": PR_id})

	sql, args, err := query.ToSql()
	if err != nil {
		return false, fmt.Errorf("[PRRepo | IsMerged]: %w", err)
	}

	err = prp.Db.QueryRow(ctx, sql, args...).Scan(&status)
	if err != nil {
		return false, fmt.Errorf("[PRRepo | IsMerged]: %w", err)
	}

	if status == string(enteties.PullRequestStatusMerged) {
		return true, nil
	} else {
		return false, nil
	}
}

func (prp *prPostgresRepository) MergePR(ctx context.Context, PR_id string) (*enteties.PullRequest, error) {

	// получим транзакцию из контекста
	tx, ok := GetTx(ctx)
	if !ok {
		return nil, fmt.Errorf("[PRRepo | MergePR]: can not get pgx.Tx")
	}

	query := prp.sq.Update("pull_requests").
		Set("status", enteties.PullRequestStatusMerged).
		Set("merged_at", squirrel.Expr("COALESCE(merged_at, NOW())")).
		Where(squirrel.Eq{"pull_request_id": PR_id})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("[PRRepo | MergePR]: %w", err)
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("[PRRepo | MergePR]: %w", err)
	}

	pr, err := prp.GetPR(ctx, PR_id)
	if err != nil {
		return nil, fmt.Errorf("[PRRepo | MergePR]: %w", err)
	}

	return pr, nil
}

func (prp *prPostgresRepository) GetPR(ctx context.Context, PR_id string) (*enteties.PullRequest, error) {

	var responcePR enteties.PullRequest

	// получим транзакцию из контекста
	tx, ok := GetTx(ctx)
	if !ok {
		return nil, fmt.Errorf("[PRRepo | GetPR]: can not get pgx.Tx")
	}

	query := prp.sq.Select("pull_request_id", "pull_request_name", "author_id", "status", "merged_at").
		From("pull_requests").
		Where(squirrel.Eq{"pull_request_id": PR_id})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("[PRRepo | GetPR]: %w", err)
	}

	row := tx.QueryRow(ctx, sql, args...)

	err = row.Scan(&responcePR.PullRequestID, &responcePR.PulRequestName, &responcePR.AuthorID,
		&responcePR.Status, &responcePR.MergedAt)
	if err != nil {
		return nil, fmt.Errorf("[PRRepo | GetPR]: %w", err)
	}

	reviewersID, err := prp.getListReviewersID(ctx, PR_id)
	if err != nil {
		return nil, fmt.Errorf("[PRRepo | GetPR]: %w", err)
	}

	responcePR.AssignedReviewers = reviewersID

	return &responcePR, nil
}

// вспомогательный метод для получения ревьюеров
func (prp *prPostgresRepository) getListReviewersID(ctx context.Context, PR_id string) ([]string, error) {
	var result []string

	// получим транзакцию из контекста
	tx, ok := GetTx(ctx)
	if !ok {
		return nil, fmt.Errorf("[PRRepo | getListReviewersID]: can not get pgx.Tx")
	}

	query := prp.sq.Select("user_id").
		From("assigned_reviewers").
		Where(squirrel.Eq{"pull_request_id": PR_id})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("[PRRepo | getListReviewersID]: %w", err)
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("[PRRepo | getListReviewersID]: %w", err)
	}

	for rows.Next() {
		var user_id string
		err := rows.Scan(&user_id)
		if err != nil {
			return nil, fmt.Errorf("[PRRepo | getListReviewersID]: %w", err)
		}

		result = append(result, user_id)
	}

	return result, nil
}

func (prp *prPostgresRepository) GetAllPRByUserID(ctx context.Context, user_id string) ([]*enteties.PullRequestShort, error) {

	// получим транзакцию из контекста
	tx, ok := GetTx(ctx)
	if !ok {
		return nil, fmt.Errorf("[PRRepo | GetAllPRByUserID]: can not get pgx.Tx")
	}

	var responce []*enteties.PullRequestShort

	query := prp.sq.Select(
		"p.pull_request_id",
		"p.pull_request_name",
		"p.author_id",
		"p.status").
		From("assigned_reviewers ar").
		LeftJoin("pull_requests p ON ar.pull_request_id = p.pull_request_id").
		Where(squirrel.Eq{"ar.user_id": user_id})

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("[PRRepo | GetAllPRByUserID]: %w", err)
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("[PRRepo | GetAllPRByUserID]: %w", err)
	}

	for rows.Next() {
		var shortPR enteties.PullRequestShort
		err := rows.Scan(&shortPR.PullRequestID, &shortPR.PulRequestName, &shortPR.AuthorID, &shortPR.Status)
		if err != nil {
			return nil, fmt.Errorf("[PRRepo | GetAllPRByUserID]: %w", err)
		}

		responce = append(responce, &shortPR)
	}

	return responce, nil
}

func (prp *prPostgresRepository) IsUserAssignedToPR(ctx context.Context, userID, prID string) (bool, error) {

	query := `SELECT EXISTS(SELECT 1 FROM assigned_reviewers WHERE pull_request_id = $1 AND 
	user_id = $2)`
	var isReviewed bool
	err := prp.Db.QueryRow(ctx, query, prID, userID).Scan(&isReviewed)
	if err != nil {
		return false, fmt.Errorf("[PRRepo | IsUserReviewedToPR]: %w", err)
	}
	return isReviewed, nil
}

func (prp *prPostgresRepository) ReassignReviewer(ctx context.Context, prID, oldUserID, newUserID string) error {
	// получим транзакцию из контекста
	tx, ok := GetTx(ctx)
	if !ok {
		return fmt.Errorf("[PRRepo | ReassignReviewer]: can not get pgx.Tx")
	}

	query := prp.sq.Update("assigned_reviewers").
		Set("user_id", newUserID).
		Where(squirrel.Eq{"pull_request_id": prID}).
		Where(squirrel.Eq{"user_id": oldUserID})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("[PRRepo | ReassignReviewer]: %w", err)
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("[PRRepo | ReassignReviewer]: %w", err)
	}

	return nil
}

func (prp *prPostgresRepository) GetAuthorPR(ctx context.Context, prID string) (string, error) {
	query := prp.sq.Select("author_id").
		From("pull_requests").
		Where(squirrel.Eq{"pull_request_id": prID})

	sql, args, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("[PRRepo | GetAuthorPR]: %w", err)
	}

	var author string

	row := prp.Db.QueryRow(ctx, sql, args...)

	err = row.Scan(&author)
	if err != nil {
		return "", fmt.Errorf("[PRRepo | GetAuthorPR]: %w", err)
	}
	return author, nil
}
