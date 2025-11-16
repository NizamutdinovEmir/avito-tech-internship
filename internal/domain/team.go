package domain

// Team represents a team with its members
type Team struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}
