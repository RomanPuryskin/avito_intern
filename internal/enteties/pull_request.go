package enteties

import "time"

type PullRequestStatus string

const (
	PullRequestStatusOpen   PullRequestStatus = "OPEN"
	PullRequestStatusMerged PullRequestStatus = "MERGED"
)

// модель описывает полную сущность pull request
type PullRequest struct {
	PullRequestID     string            `json:"pull_request_id"`
	PulRequestName    string            `json:"pull_request_name"`
	AuthorID          string            `json:"author_id"`
	Status            PullRequestStatus `json:"status"`
	AssignedReviewers []string          `json:"assigned_reviewers"` // (user_id)
	CreatedAt         *time.Time        `json:"created_at,omitempty"`
	MergedAt          *time.Time        `json:"merged_at,omitempty"`
}

// модель описывает упрощенную сущность pull request
type PullRequestShort struct {
	PullRequestID  string            `json:"pull_request_id"`
	PulRequestName string            `json:"pull_request_name"`
	AuthorID       string            `json:"author_id"`
	Status         PullRequestStatus `json:"status"`
}

// модель описывает формат запроса на создание pull request
type CreatePullRequest struct {
	PullRequestID   string `json:"pull_request_id" validate:"required"`
	PullRequestName string `json:"pull_request_name" validate:"required"`
	AuthorID        string `json:"author_id" validate:"required"`
}

// модель описывает формат запроса на мердж pull request
type MergePullRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
}

// модель описывает формат запроса на переназанчение ревьюера на pull request
type ReassignPullRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
	OldUserID     string `json:"old_user_id" validate:"required"`
}

// модель описывает формат ответа на запрос о переназначении ревьюера на pull request
type ReassignPullRequestResponce struct {
	PR         PullRequest `json:"pr"`
	ReplacedBy string      `json:"replaced_by"`
}
