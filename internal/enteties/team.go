package enteties

// модель описывает полную сущность команды (team)
type Team struct {
	TeamName string       `json:"team_name" validate:"required"`
	Members  []TeamMember `json:"members" validate:"required"`
}
