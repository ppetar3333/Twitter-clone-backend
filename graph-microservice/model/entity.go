package model

type Entity struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Type     string `json:"type" validate:"required"`
}
