package model

type UserBusiness struct {
	Id       string `json:"id" bson:"id"`
	Username string `json:"username" bson:"username"`
	Website  string `json:"website" bson:"website"`
	Email    string `json:"email" bson:"email"`
	Company  string `json:"company" bson:"company"`
	Role     string `json:"role" bson:"role"`
	PrivateProfile bool   `json:"privateProfile" bson:"privateProfile"`
}
