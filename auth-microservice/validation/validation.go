package validation

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/ppetar33/twitter-api/model"
	database "github.com/ppetar33/twitter-api/server"
)

func SignUpValidationRegularUser(user *model.User) string {

	var errText string

	var hasUpper bool
	var hasLower bool

	if (user.Username == "") || (user.Password == "") || (user.FirstName == "") || (user.LastName == "") || (user.Gender == "") || (user.Age == "") || (user.City == "") {
		errText = `All fields are required!`
		return errText
	}

	isValid, errText := IsPasswordInBlackList(user.Password)
	if !isValid {
		return errText
	}

	for _, r := range user.Password {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
	}

	numeric := regexp.MustCompile(`\d`).MatchString(user.Password)

	if len(user.Password) < 8 || !numeric || !hasUpper || !hasLower {
		errText = "Minimum eight characters, at least one uppercase letter, one lowercase letter and one number!"
		return errText
	}

	for _, r := range user.FirstName {
		if !unicode.IsLetter(r) {
			errText = "First name can contain only letters!"
			return errText
		}
	}

	for _, r := range user.LastName {
		if !unicode.IsLetter(r) {
			errText = "Last name can contain only letters!"
			return errText
		}
	}

	for _, r := range user.City {
		if !unicode.IsLetter(r) {
			errText = "City can contain only letters!"
			return errText
		}
	}

	if strings.TrimRight(user.Gender, "\n") != "male" && strings.TrimRight(user.Gender, "\n") != "female" {
		errText = "Gender can only be male or female!"
		return errText
	}

	if _, err := strconv.Atoi(user.Age); err != nil {
		errText = "Age can only contain digit!"
		return errText
	}

	if (user.Email != "") && !isEmailValid(user.Email) {
		errText = `Email must be like example@gmail.com!`
		return errText
	}

	return errText
}

func SignUpValidationBusinessUser(user *model.UserBusiness) string {

	var errText string

	var hasUpper bool
	var hasLower bool

	fmt.Print("USERNAME BUSINESS " + user.Username)

	userExists := database.GetUserByUsername(user.Username)
	if userExists != (model.Auth{}) {
		errText = `Username Exists!`
		return errText
	}

	userExists1 := database.IsUserWithEmailExist(user.Email)
	if userExists1 != (model.Auth{}) {
		errText = `Email Exists!`
		return errText
	}

	if (user.Username == "") || (user.Website == "") || (user.Email == "") || (user.Company == "") {
		errText = `All fields is required!`
		return errText
	}

	isValid, errText := IsPasswordInBlackList(user.Password)
	if !isValid {
		return errText
	}

	for _, r := range user.Password {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
	}

	numeric := regexp.MustCompile(`\d`).MatchString(user.Password)

	if len(user.Password) < 8 || !numeric || !hasUpper || !hasLower {
		errText = "Minimum eight characters, at least one uppercase letter, one lowercase letter and one number!"
		return errText
	}

	if (user.Email != "") && !isEmailValid(user.Email) {
		errText = `Email must be like example@gmail.com!`
		return errText
	}

	return errText
}

func isEmailValid(e string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(e)
}

func EmailValidation(codeRecovery model.CodeRecovery) (bool, string) {
	isValid := true
	var errText string

	if (codeRecovery.Email != "") && !isEmailValid(codeRecovery.Email) {
		errText = `Email must be like example@gmail.com!`
		isValid = false
		return isValid, errText
	}

	return isValid, errText
}

func IsChangePasswordValid(pass model.ChangePassword) (bool, string) {
	isValid := true
	var errorMessage string

	if (pass.NewPassword == "") || (pass.CurrentPassword == "") {
		errorMessage = `All fields is required!`
		isValid = false
		return isValid, errorMessage
	}

	isValid, errText := IsPasswordInBlackList(pass.NewPassword)
	if !isValid {
		isValid = false
		return isValid, errText
	}

	var hasUpper bool
	var hasLower bool

	for _, r := range pass.NewPassword {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
	}

	numeric := regexp.MustCompile(`\d`).MatchString(pass.NewPassword)

	if len(pass.NewPassword) < 8 || !numeric || !hasUpper || !hasLower {
		isValid = false
		errorMessage = "Minimum eight characters, at least one uppercase letter, one lowercase letter and one number"
		return isValid, errorMessage
	}

	return isValid, errorMessage
}

func IsRecoveryPasswordValid(pass model.RecoveryPassword) (bool, string) {
	isValid := true
	var errorMessage string

	if (pass.NewPassword == "") || (pass.Email == "") || (pass.Code == "") {
		errorMessage = `All fields is required!`
		isValid = false
		return isValid, errorMessage
	}

	isValid, errText := IsPasswordInBlackList(pass.NewPassword)
	if !isValid {
		isValid = false
		return isValid, errText
	}

	var hasUpper bool
	var hasLower bool

	for _, r := range pass.NewPassword {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
	}

	numeric := regexp.MustCompile(`\d`).MatchString(pass.NewPassword)

	if len(pass.NewPassword) < 8 || !numeric || !hasUpper || !hasLower {
		isValid = false
		errorMessage = "Minimum eight characters, at least one uppercase letter, one lowercase letter and one number"
		return isValid, errorMessage
	}

	if (pass.Email != "") && !isEmailValid(pass.Email) {
		errText = `Email must be like example@gmail.com!`
		isValid = false
		return isValid, errText
	}

	return isValid, errorMessage
}

func IsPasswordInBlackList(password string) (bool, string) {
	isValid := true
	var errorMessage string

	f, err := os.Open("blacklist.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		if password == scanner.Text() {
			isValid = false
			errorMessage = "Use stronger password!"
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return isValid, errorMessage
}
