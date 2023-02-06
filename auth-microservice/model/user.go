package model

type User struct {
	Id        string `json:"id" bson:"id"`
	FirstName string `json:"firstname" bson:"firstname"`
	LastName  string `json:"lastname" bson:"lastname"`
	Password  string `json:"password" bson:"password"`
	Username  string `json:"username" bson:"username"`
	Email     string `json:"email" bson:"email"`
	Gender    string `json:"gender" bson:"gender"`
	Age       string `json:"age" bson:"age"`
	City      string `json:"city" bson:"city"`
	Role      string `json:"role" bson:"role"`
}
