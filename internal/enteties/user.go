package enteties

// модель описывает полную сущность пользователя (User)
type User struct {
	UserID   string `json:"user_id"`
	UserName string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

// модель описывает формат запроса для изменения статуса пользователя
type RequestUserToSetActive struct {
	UserID   string `json:"user_id" validate:"required"`
	IsActive bool   `json:"is_active" validate:"required"`
}

// модель описывает члена команды
type TeamMember struct {
	UserID   string `json:"user_id" validate:"required"`
	UserName string `json:"username" validate:"required"`
	IsActive bool   `json:"is_active" validate:"required"`
}

// модель описывает формат запроса на получение всех pull request, на которые пользователь
// назначение ревьюером
type UserReviews struct {
	UserID       string             `json:"user_id"`
	PullRequests []PullRequestShort `json:"pull_requests"`
}
