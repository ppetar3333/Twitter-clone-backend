package main

import (
	"encoding/json"
	"io"
	"regexp"
	"strconv"

	"github.com/ppetar33/twitter-api/profile-microservice/model"
)

func DecodeBodyRegister(r io.Reader) (*model.User, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt model.User
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}

	return &rt, nil
}

func DecodeBodyRegisterBusinessUser(r io.Reader) (*model.UserBusiness, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt model.UserBusiness
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}

	return &rt, nil
}

func DecodeBodyListOfIDs(r io.Reader) (*[]string, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt []string
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}

	return &rt, nil
}

func SignUpValidationRegularUser(user *model.User) string {

	errText := ""

	if (user.Username == "") || (user.FirstName == "") || (user.LastName == "") || (user.Gender == "") || (user.Age == "") || (user.City == "") {
		errText = "All fields is required!"
	}

	if _, err := strconv.Atoi(user.Age); err != nil {
		errText = "Age must be number!"
	}

	if !isEmailValid(user.Email) {
		errText = "Email must be like example@gmail.com!"
	}

	return errText
}

func SignUpValidationBusinessUser(user *model.UserBusiness) string {

	errText := ""

	if (user.Username == "") || (user.Website == "") || (user.Email == "") || (user.Company == "") {
		errText = "All fields is required!"
	}

	if !isEmailValid(user.Email) {
		errText = "Email must be like example@gmail.com!"
	}

	return errText
}

func isEmailValid(e string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(e)
}
