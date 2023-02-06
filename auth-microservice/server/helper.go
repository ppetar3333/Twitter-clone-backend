package server

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"os"
	"time"
)

func GenerateKey() (string, string) {
	id := uuid.New().String()
	return fmt.Sprintf(id), id
}

func GetHash(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		fmt.Println(err)
	}
	return string(hash)
}

var SECRET_KEY = []byte("gosecretkey")

func GenerateJWT(username string, role string) (string, error) {
	var err error
	os.Setenv("ACCESS_SECRET", string(SECRET_KEY))
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["username"] = username
	atClaims["role"] = role
	atClaims["exp"] = time.Now().Add(time.Minute * 15).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(os.Getenv("ACCESS_SECRET")))
	if err != nil {
		return "", err
	}
	return token, nil
}

func GenerateRandomSaltNumber() int {
	min := 0
	max := 1000
	return rand.Intn(max - min)
}
