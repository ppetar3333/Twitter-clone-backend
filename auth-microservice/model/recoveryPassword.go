package model

type RecoveryPassword struct {
	Email       string `json:"email" bson:"email"`
	Code        string `json:"code" bson:"code"`
	NewPassword string `json:"newPassword" bson:"newPassword"`
}
