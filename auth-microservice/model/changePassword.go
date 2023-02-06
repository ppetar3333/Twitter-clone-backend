package model

type ChangePassword struct {
	CurrentPassword string `json:"currentPassword" bson:"currentPassword"`
	NewPassword     string `json:"newPassword" bson:"newPassword"`
}
