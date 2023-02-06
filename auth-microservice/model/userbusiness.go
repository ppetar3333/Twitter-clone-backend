package model

type UserBusiness struct {
	Id       string `json:"id" bson:"id"`
	Password string `json:"password" bson:"password"`
	Username string `json:"username" bson:"username"`
	Website  string `json:"website" bson:"website"`
	Email    string `json:"email" bson:"email"`
	Company  string `json:"company" bson:"company"`
	Role     string `json:"role" bson:"role"`
}
