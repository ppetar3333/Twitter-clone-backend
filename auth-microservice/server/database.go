package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/ppetar33/twitter-api/model"
	"github.com/sony/gobreaker"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

var client *mongo.Client
var cb = CircuitBreaker()

const applicationJson = "application/json"
const errorStatus = "error status: "
const errorOccured = "An Error Occured %v"
const errorMessage = "Error occured "

type AuthRepository struct {
	Tracer trace.Tracer
}

func ConnectToAuthDatabase() (*mongo.Client, error) {
	ctx, err := context.WithTimeout(context.Background(), 10*time.Second)
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://auth-db:27017"))

	defer err()

	if err == nil {
		fmt.Println("Error with database")
	}

	return client, nil
}

func Login(user *model.Auth) (string, error) {
	var dbUser model.Auth

	collection := client.Database("AUTH").Collection("user")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOne(ctx, bson.M{"username": user.Username}).Decode(&dbUser)

	defer cancel()

	if err != nil {
		fmt.Println("Username error")
		return "Wrong Credentials, please check all fields!", nil
	}

	if dbUser.Status != "verified" {
		return "Your profile is not verified", nil
	}

	userPass := []byte(user.Password + dbUser.Salt)
	dbPass := []byte(dbUser.Password)

	passErr := bcrypt.CompareHashAndPassword(dbPass, userPass)

	if passErr != nil {
		fmt.Println(passErr)
		return "Wrong Credentials, please check all fields!", nil
	}

	jwtToken, errToken := GenerateJWT(user.Username, dbUser.Role)
	if errToken != nil {
		return `{"message":"` + errToken.Error() + `"}`, nil
	}

	return jwtToken, nil
}

func ValidateCode(user *model.Auth) (string, error) {
	var dbUser model.Auth

	collection := client.Database("AUTH").Collection("user")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOne(ctx, bson.M{"username": user.Username}).Decode(&dbUser)

	defer cancel()

	if err != nil {
		fmt.Println("Username error")
		return "Wrong Username Or Password!", nil
	}

	userPass := []byte(user.Password + dbUser.Salt)
	dbPass := []byte(dbUser.Password)

	passErr := bcrypt.CompareHashAndPassword(dbPass, userPass)

	if passErr != nil {
		fmt.Println(passErr)
		return "Wrong Username Or Password!", nil
	}

	if dbUser.Status != "verified" {
		if user.Code == dbUser.Code {
			_, err1 := collection.UpdateOne(
				ctx,
				bson.M{"username": user.Username},
				bson.D{
					{"$set", bson.D{{"status", "verified"}}},
				},
			)
			if err1 != nil {
				log.Fatal(err)
			}
			return "Code is valid, you are verified", nil
		} else {
			return "Code is not valid", nil
		}
	}

	return "", nil
}

func (r AuthRepository) SaveCredentialsRegularIntoAuth(ctx context.Context, user *model.User) (*mongo.InsertOneResult, error) {
	ctx, span := r.Tracer.Start(ctx, "AuthRepository.SaveCredentialsRegularIntoAuth")
	defer span.End()
	salt := GenerateRandomSaltNumber()

	user.Id = uuid.New().String()
	user.Password = GetHash([]byte(user.Password + strconv.Itoa(salt)))
	user.Role = "regular"

	userExists := GetUserByUsername(user.Username)
	if userExists != (model.Auth{}) {
		return nil, errors.New("Username Exists!")
	}

	userExists1 := IsUserWithEmailExist(user.Email)
	if userExists1 != (model.Auth{}) {
		return nil, errors.New("User with this email exist!")
	}

	authResult, err := SaveCredentialsIntoAuth(user.Username, user.Password, user.Role, strconv.Itoa(salt), user.Email)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		fmt.Println(err)
	}

	span.SetStatus(codes.Ok, "")

	errProfile := r.SaveRegularUserIntoProfile(ctx, user)
	if errProfile != nil {
		profileError := r.CancelSaveIntoProfile(ctx, user)
		if profileError != nil {
			return nil, profileError
		}
		return nil, errors.New("profile - user cannot be created")
	}

	errGraph := r.SaveRegularUserIntoGraph(ctx, user)
	if errGraph != nil {
		graphError := r.CancelSaveIntoGraph(ctx, user)
		if graphError != nil {
			return nil, graphError
		}
		return nil, errors.New("graph - user cannot be created")
	}

	return authResult, nil
}

func SaveCredentialsBusinessIntoAuth(user *model.UserBusiness) (*mongo.InsertOneResult, error) {
	salt := GenerateRandomSaltNumber()
	user.Password = GetHash([]byte(user.Password + strconv.Itoa(salt)))
	user.Role = "business"

	authResult, err := SaveCredentialsIntoAuth(user.Username, user.Password, user.Role, strconv.Itoa(salt), user.Email)

	if err != nil {
		fmt.Println(err)
	}

	errProfile := SaveBusinessUserIntoProfile(user)
	if errProfile != nil {
		// profileError :=	CancelSaveIntoProfile(user)
		// if profileError != nil {
		// 	return nil, profileError
		// }
		// return nil, errors.New("profile - user cannot be created")

		fmt.Print("Error profile: ", errProfile.Error())
		return nil, errProfile
	}

	errGraph := SaveBusinessUserIntoGraph(user)
	if errGraph != nil {
		// graphError := CancelSaveIntoGraph(user)
		// if graphError != nil {
		// 	return nil, graphError
		// }
		// return nil, errors.New("graph - user cannot be created")

		fmt.Print("Error graph: ", errGraph.Error())
		return nil, errGraph
	}

	return authResult, nil
}

func SaveCredentialsIntoAuth(username string, password string, role string, salt string, email string) (*mongo.InsertOneResult, error) {
	collection := client.Database("AUTH").Collection("user")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var result model.Auth

	userErr := collection.FindOne(ctx, bson.M{"username": username}).Decode(&result)

	fmt.Print("cannot find user with username" + userErr.Error())

	if result.Username == "" && userErr != nil {
		var auth model.Auth

		min := 100000
		max := 999999
		rand.Seed(time.Now().UnixNano())
		code := rand.Intn(max-min) + min

		auth.Username = username
		auth.Password = password
		auth.Email = email
		auth.Salt = salt
		auth.Role = role
		auth.Status = "pending"
		auth.Code = strconv.Itoa(code)

		from := "tim.osam.twitter.clone@gmail.com"
		sifra := "bxafykgbwdgfzqre"

		to := []string{
			auth.Email,
		}

		smtpHost := "smtp.gmail.com"
		smtpPort := "587"

		msg := []byte("Subject: Your validation code is here!\r\n" +
			"\r\n" +
			"This is your validation code. \r\n " + strconv.Itoa(code))

		au := smtp.PlainAuth("", from, sifra, smtpHost)
		err := smtp.SendMail(smtpHost+":"+smtpPort, au, from, to, msg)
		if err != nil {
			fmt.Println("ERROR")
			fmt.Println(err)
			return nil, nil
		}
		fmt.Println("Email Sent Successfully!")

		resultAuth, _ := collection.InsertOne(ctx, auth)

		fmt.Println("****** Result inserted into AUTH database   ******  : ", auth)

		return resultAuth, nil
	} else {
		fmt.Println(`Username Exists!`)
		return nil, nil
	}
}

func (r AuthRepository) SaveRegularUserIntoProfile(ctx context.Context, user *model.User) error {
	ctx, span := r.Tracer.Start(ctx, "AuthRepository.SaveRegularUserIntoProfile")
	defer span.End()

	usr := map[string]string{
		"id":        user.Id,
		"username":  user.Username,
		"email":     user.Email,
		"firstname": user.FirstName,
		"lastname":  user.LastName,
		"age":       user.Age,
		"city":      user.City,
		"gender":    user.Gender,
		"role":      user.Role,
	}

	userToSend, _ := json.Marshal(usr)

	fmt.Println("user regular", usr)

	responseBody := bytes.NewBuffer(userToSend)

	respProfile, errProfile := cb.Execute(func() (interface{}, error) {
		resp, err := http.Post("http://profile-microservice:8081/api/profile/regular-user-sign-up", applicationJson, responseBody)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New(errorStatus + resp.Status)
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return bodyBytes, nil
	})

	if errProfile != nil {
		fmt.Println(errorOccured, errProfile)
		span.SetStatus(codes.Error, errProfile.Error())
		return errors.New("An Error Prfoile Occured: " + errProfile.Error())
	}

	span.SetStatus(codes.Ok, "")
	fmt.Println(string(respProfile.([]byte)))

	return nil
}

func (r AuthRepository) CancelSaveIntoProfile(ctx context.Context, user *model.User) error {
	ctx, span := r.Tracer.Start(ctx, "AuthRepository.CancelSaveIntoProfile")
	defer span.End()
	client := &http.Client{}

	path := "http://profile-microservice:8081/api/profile/cancel-user-register/" + user.Id
	reqProfile, errProfile := http.NewRequest("DELETE", path, nil)

	if errProfile != nil {
		fmt.Println(errProfile)
		return errors.New(errProfile.Error())
	}

	resp, err := client.Do(reqProfile)
	if err != nil {
		fmt.Println(err)
		return errors.New(err.Error())
	}

	defer resp.Body.Close()

	respBody, errResp := ioutil.ReadAll(resp.Body)
	if errResp != nil {
		fmt.Println(errResp)
		return errors.New(errResp.Error())
	}

	fmt.Println("response Status : ", resp.Status)
	fmt.Println("response Headers : ", resp.Header)
	fmt.Println("response Body : ", string(respBody))

	return nil
}

func (r AuthRepository) CancelSaveIntoGraph(ctx context.Context, user *model.User) error {
	ctx, span := r.Tracer.Start(ctx, "AuthRepository.CancelSaveIntoGraph")
	defer span.End()
	client := &http.Client{}

	path := "http://profile-microservice:8081/api/socialGraph/deleteUser/" + user.Id
	reqGraph, errGraph := http.NewRequest("DELETE", path, nil)

	if errGraph != nil {
		fmt.Println(errGraph)
		return errors.New(errGraph.Error())
	}

	resp, err := client.Do(reqGraph)
	if err != nil {
		fmt.Println(err)
		return errors.New(err.Error())
	}

	defer resp.Body.Close()

	respBody, errResp := ioutil.ReadAll(resp.Body)
	if errResp != nil {
		fmt.Println(errResp)
		return errors.New(errResp.Error())
	}

	fmt.Println("response Status : ", resp.Status)
	fmt.Println("response Headers : ", resp.Header)
	fmt.Println("response Body : ", string(respBody))

	return nil
}

func SaveBusinessUserIntoProfile(user *model.UserBusiness) error {

	user.Id = uuid.New().String()
	user.Role = "business"

	usr := map[string]string{
		"id":       user.Id,
		"username": user.Username,
		"company":  user.Company,
		"website":  user.Website,
		"email":    user.Email,
		"role":     user.Role,
	}

	fmt.Println("user business", usr)

	userToSend, _ := json.Marshal(usr)

	responseBody := bytes.NewBuffer(userToSend)

	respProfile, errProfile := cb.Execute(func() (interface{}, error) {
		resp, err := http.Post("http://profile-microservice:8081/api/profile/business-user-sign-up", applicationJson, responseBody)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New(errorStatus + resp.Status)
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return bodyBytes, nil
	})

	if errProfile != nil {
		fmt.Println(errorOccured, errProfile)
		return errors.New("An Error Prfoile Occured: " + errProfile.Error())
	}

	fmt.Println(string(respProfile.([]byte)))

	return nil
}

func (r AuthRepository) SaveRegularUserIntoGraph(ctx context.Context, user *model.User) error {
	ctx, span := r.Tracer.Start(ctx, "AuthRepository.SaveRegularUserIntoGraph")
	defer span.End()

	usr := map[string]string{
		"id":       user.Id,
		"type":     user.Role,
		"name":     user.FirstName + " " + user.LastName,
		"username": user.Username,
	}

	fmt.Println("user regular", usr)

	userToSend, _ := json.Marshal(usr)

	responseBody := bytes.NewBuffer(userToSend)

	respGraph, errGraph := cb.Execute(func() (interface{}, error) {
		resp, err := http.Post("http://graph-microservice:8082/api/socialGraph/registerUser", applicationJson, responseBody)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New(errorStatus + resp.Status)
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return bodyBytes, nil
	})

	if errGraph != nil {
		fmt.Println(errorOccured, errGraph)
		span.SetStatus(codes.Error, errGraph.Error())
		return errors.New("An Error Graph Occured: " + errGraph.Error())
	}

	span.SetStatus(codes.Ok, "")
	fmt.Println(string(respGraph.([]byte)))
	return nil
}

func SaveBusinessUserIntoGraph(user *model.UserBusiness) error {

	usr := map[string]string{
		"id":       user.Id,
		"type":     user.Role,
		"name":     user.Company,
		"username": user.Username,
	}

	fmt.Println("user business", usr)

	userToSend, _ := json.Marshal(usr)

	responseBody := bytes.NewBuffer(userToSend)

	respGraph, errGraph := cb.Execute(func() (interface{}, error) {
		resp, err := http.Post("http://graph-microservice:8082/api/socialGraph/registerUser", applicationJson, responseBody)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New(errorStatus + resp.Status)
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return bodyBytes, nil
	})

	if errGraph != nil {
		fmt.Println(errorOccured, errGraph)
		return errors.New("An Error Graph Occured: " + errGraph.Error())
	}

	fmt.Println(string(respGraph.([]byte)))

	return nil
}

func Disconnect(ctx context.Context) error {
	err := client.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
}

func ChangePassword(password model.ChangePassword, username string) (error, string) {

	var errorString string
	var dbUser model.Auth

	collection := client.Database("AUTH").Collection("user")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&dbUser)
	if err != nil {
		errorString = "Doesn't exist user with this username"
	}

	enteredPass := []byte(password.CurrentPassword + dbUser.Salt)
	dbPass := []byte(dbUser.Password)

	passErr := bcrypt.CompareHashAndPassword(dbPass, enteredPass)
	if passErr != nil {
		errorString = "The password entered is incorrect"
	} else {
		newPassword := GetHash([]byte(password.NewPassword + dbUser.Salt))

		_, err1 := collection.UpdateOne(
			ctx,
			bson.M{"username": username},
			bson.D{
				{"$set", bson.D{{"password", newPassword}}},
			},
		)
		if err1 != nil {
			log.Fatal(err)
		}
	}

	return nil, errorString
}

func RecoveryPassword(recovery model.RecoveryPassword) (error, string) {
	var errorString string
	var dbRecovery model.EmailCodeRecovery

	collection := client.Database("AUTH").Collection("passwordRecovery")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	errEmail := collection.FindOne(ctx, bson.M{"email": recovery.Email}).Decode(&dbRecovery)

	if errEmail != nil {
		log.Println("Please check your email!")
		errorString = "Please check your email!"
		return errors.New(errEmail.Error()), errorString
	}

	errCode := collection.FindOne(ctx, bson.M{"code": recovery.Code}).Decode(&dbRecovery)

	if errCode != nil {
		log.Println("Please check your code!")
		errorString = "Please check your code!"
		return errors.New(errCode.Error()), errorString
	}

	var dbUser model.Auth

	collectionUser := client.Database("AUTH").Collection("user")
	ctxUser, _ := context.WithTimeout(context.Background(), 10*time.Second)
	errUser := collectionUser.FindOne(ctxUser, bson.M{"email": recovery.Email}).Decode(&dbUser)

	if errUser != nil {
		log.Println("Cannot find user!")
		errorString = "Cannot find user!"
		return errors.New(errUser.Error()), errorString
	}

	newPassword := GetHash([]byte(recovery.NewPassword + dbUser.Salt))

	_, errUpdate := collectionUser.UpdateOne(
		ctx,
		bson.M{"email": dbUser.Email},
		bson.D{
			{"$set", bson.D{{"password", newPassword}}},
		},
	)

	if errUpdate != nil {
		log.Fatal(errUpdate)
		return errors.New(errUpdate.Error()), errUpdate.Error()
	}

	return nil, ""
}

// Check database connection
func Ping() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection -> if no error, connection is established
	err := client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
	}

	// Print available databases
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(databases)
}

func CircuitBreaker() *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(
		gobreaker.Settings{
			Name:        "cb",
			MaxRequests: 3,
			Timeout:     2 * time.Second,
			Interval:    0,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 0
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				fmt.Printf("Circuit Breaker '%s' changed from '%s' to '%s'\n", name, from, to)
			},
		},
	)
}

func GetUserByUsername(username string) model.Auth {
	collection := client.Database("AUTH").Collection("user")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var user model.Auth

	err := collection.FindOne(ctx, bson.D{{"username", username}}).Decode(&user)

	if err != nil {
		fmt.Println(errorMessage + err.Error())
		return user
	}

	return user
}

func GetUserByEmail(codeRecovery model.CodeRecovery) model.Auth {
	collection := client.Database("AUTH").Collection("user")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var user model.Auth

	fmt.Print(codeRecovery.Email)
	err := collection.FindOne(ctx, bson.D{{"email", codeRecovery.Email}}).Decode(&user)

	fmt.Print(user)
	if err != nil {
		fmt.Println(errorMessage + err.Error())
		return user
	}

	return user
}

func IsUserWithEmailExist(email string) model.Auth {
	collection := client.Database("AUTH").Collection("user")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var user model.Auth

	err := collection.FindOne(ctx, bson.D{{"email", email}}).Decode(&user)

	if err != nil {
		fmt.Println(errorMessage + err.Error())
		return user
	}

	return user
}

func SendCodeViaEmail(codeRecovery model.CodeRecovery) (*mongo.InsertOneResult, string) {
	collection := client.Database("AUTH").Collection("passwordRecovery")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var emailCode model.EmailCodeRecovery

	min := 100000
	max := 999999
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(max-min) + min

	emailCode.Email = codeRecovery.Email
	emailCode.Code = strconv.Itoa(code)

	resultAuth, _ := collection.InsertOne(ctx, emailCode)

	from := "tim.osam.twitter.clone@gmail.com"
	password := os.Getenv("EMAIL_PASSWORD")

	to := []string{
		codeRecovery.Email,
	}

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	msg := []byte("Subject: Your code is here!\r\n" +
		"\r\n" +
		"This is your code for recovery password. \r\n " + strconv.Itoa(code))

	au := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, au, from, to, msg)
	if err != nil {
		fmt.Println("ERROR")
		fmt.Println(err)
		return nil, "Error sending email"
	}
	fmt.Println("Email Sent Successfully!")

	return resultAuth, ""
}
