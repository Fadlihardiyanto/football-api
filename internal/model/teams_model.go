package model

type TeamResponse struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	Logo                string            `json:"logo"`
	FoundedYear         int               `json:"founded_year"`
	HeadquartersAddress string            `json:"headquarters_address"`
	HeadquartersCity    string            `json:"headquarters_city"`
	CreatedAt           string            `json:"created_at"`
	UpdatedAt           string            `json:"updated_at"`
	DeletedAt           *string           `json:"deleted_at,omitempty"`
	Players             []*PlayerResponse `json:"players,omitempty"`
}

type TeamRequestCreate struct {
	Name                string `json:"name" validate:"required"`
	Logo                string `json:"logo" validate:"omitempty"`
	FoundedYear         int    `json:"founded_year" validate:"required"`
	HeadquartersAddress string `json:"headquarters_address" validate:"required"`
	HeadquartersCity    string `json:"headquarters_city" validate:"required"`
}

type TeamRequestUpdate struct {
	ID                  string `json:"id" validate:"required,uuid"`
	Name                string `json:"name" validate:"omitempty"`
	Logo                string `json:"logo" validate:"omitempty"`
	FoundedYear         int    `json:"founded_year" validate:"omitempty"`
	HeadquartersAddress string `json:"headquarters_address" validate:"omitempty"`
	HeadquartersCity    string `json:"headquarters_city" validate:"omitempty"`
}

type TeamRequestFindByID struct {
	ID string `json:"id" validate:"required,uuid"`
}

type TeamRequestSoftDelete struct {
	ID string `json:"id" validate:"required,uuid"`
}

type TeamRequestUploadLogo struct {
	ID   string `json:"id" validate:"required,uuid"`
	Logo string `json:"logo" validate:"required"`
}
