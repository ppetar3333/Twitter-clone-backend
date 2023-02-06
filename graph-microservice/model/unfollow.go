package model

type Unfollow struct {
	User           string `json:"user"`
	WantToUnfollow string `json:"wantToUnfollow"`
}
