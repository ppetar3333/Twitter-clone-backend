package model

import "time"

type Tweet struct {
	Id        string    `json:"id"`
	Text      string    `json:"text"`
	Likes     *[]string `json:"likes"`
	Tweet     string    `json:"tweet"`
	User      string    `json:"user"`
	CreatedAt time.Time `json:"createdAt"`
}
