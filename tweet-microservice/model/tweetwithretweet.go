package model

import "time"

type TweetWithRetweet struct {
	Id    string    `json:"id"`
	Text  string    `json:"text"`
	Likes *[]string `json:"likes"`
	Tweet TweetWithUser    `json:"tweet"`
	User  UserProfile     `json:"user"`
	CreatedAt time.Time `json:"createdAt"`
}
