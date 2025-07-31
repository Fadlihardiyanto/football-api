package model

type PlayerResponse struct {
	ID           string        `json:"id"`
	TeamID       string        `json:"team_id"`
	Name         string        `json:"name"`
	Height       float64       `json:"height"`
	Weight       float64       `json:"weight"`
	Position     string        `json:"position"`
	JerseyNumber int           `json:"jersey_number"`
	CreatedAt    string        `json:"created_at"`
	UpdatedAt    string        `json:"updated_at"`
	DeletedAt    *string       `json:"deleted_at,omitempty"`
	Team         *TeamResponse `json:"team"`
}

type PlayerRequestCreate struct {
	Name         string  `json:"name" validate:"required"`
	TeamID       string  `json:"team_id" validate:"required,uuid"`
	Height       float64 `json:"height" validate:"required,gt=0"`
	Weight       float64 `json:"weight" validate:"required,gt=0"`
	Position     string  `json:"position" validate:"required,oneof=penyerang gelandang bertahan penjaga_gawang"`
	JerseyNumber int     `json:"jersey_number" validate:"required,min=1,max=99"`
}

type PlayerRequestUpdate struct {
	ID           string  `json:"id" validate:"required,uuid"`
	Name         string  `json:"name" validate:"omitempty"`
	TeamID       string  `json:"team_id" validate:"omitempty,uuid"`
	Height       float64 `json:"height" validate:"omitempty,gt=0"`
	Weight       float64 `json:"weight" validate:"omitempty,gt=0"`
	Position     string  `json:"position" validate:"omitempty,oneof=penyerang gelandang bertahan penjaga_gawang"`
	JerseyNumber int     `json:"jersey_number" validate:"omitempty,min=1,max=99"`
}

type PlayerRequestFindByID struct {
	ID string `json:"id" validate:"required,uuid"`
}

type PlayerRequestSoftDelete struct {
	ID string `json:"id" validate:"required,uuid"`
}
