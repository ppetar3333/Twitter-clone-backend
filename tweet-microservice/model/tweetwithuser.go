package model

import "time"

type TweetWithUser struct {
	Id    string    `json:"id"`
	Text  string    `json:"text"`
	Likes *[]string `json:"likes"`
	Tweet string    `json:"tweet"`
	User  UserProfile     `json:"user"`
	CreatedAt time.Time `json:"createdAt"`
}
