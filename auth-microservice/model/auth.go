package model

type Auth struct {
	Password string `json:"password" bson:"password"`
	Username string `json:"username" bson:"username"`
	Email    string `json:"email" bson:"email"`
	Salt     string `json:"salt" bson:"salt"`
	Role     string `json:"role" bson:"role"`
	Status   string `json:"status" bson:"status"`
	Code     string `json:"code" bson:"code"`
}
