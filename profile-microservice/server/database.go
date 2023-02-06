package server

import (
	"context"
	"fmt"
	"github.com/ppetar33/twitter-api/profile-microservice/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"log"
	"time"
)

var client *mongo.Client

type ProfileRepository struct {
	Tracer trace.Tracer
}

func ConnectToProfileDatabase() (*mongo.Client, error) {
	ctx, err := context.WithTimeout(context.Background(), 10*time.Second)
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://profile-db:27017"))

	defer err()

	if err == nil {
		fmt.Println("Error with database")
	}

	return client, nil
}

// REGULAR USER
func (r ProfileRepository) GetRegularUsers(ctxx context.Context) ([]model.User, error) {
	ctxx, span := r.Tracer.Start(ctxx, "ProfileRepository.GetRegularUsers")

	collection := client.Database("PROFILE").Collection("user")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	filter := bson.D{{"role", "regular"}}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		panic(err)
	}
	var users []model.User
	if err = cursor.All(ctx, &users); err != nil {
		span.SetStatus(codes.Error, err.Error())
		fmt.Println(err)
	}
	span.SetStatus(codes.Ok, "")
	return users, nil
}

func (r ProfileRepository) GetRegularUserByUsername(ctxx context.Context, username string) model.User {
	ctxx, span := r.Tracer.Start(ctxx, "ProfileRepository.GetRegularUserByUsername")

	collection := client.Database("PROFILE").Collection("user")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var user model.User
	err := collection.FindOne(ctx, bson.D{{"username", username}}).Decode(&user)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		fmt.Println("Doesn't exist user with this username")
	}
	span.SetStatus(codes.Ok, "")
	return user
}

func (r ProfileRepository) GetRegularUserById(ctxx context.Context, id string) model.User {

	ctxx, span := r.Tracer.Start(ctxx, "ProfileRepository.GetRegularUserById")
	collection := client.Database("PROFILE").Collection("user")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var user model.User
	err := collection.FindOne(ctx, bson.D{{"id", id}}).Decode(&user)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		fmt.Println("Doesn't exist user with this id")
	}
	span.SetStatus(codes.Ok, "")
	return user
}

func (r ProfileRepository) SignUpRegularUser(ctxx context.Context, user *model.User) (*mongo.InsertOneResult, error) {
	ctxx, span := r.Tracer.Start(ctxx, "ProfileRepository.SignUpRegularUser")
	collection := client.Database("PROFILE").Collection("user")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	user.PrivateProfile = true

	result, err := collection.InsertOne(ctx, user)
	span.SetStatus(codes.Ok, "")
	return result, err
}

//BUSINESS USER

func (r ProfileRepository) GetBusinessUsers(ctxx context.Context) ([]model.UserBusiness, error) {
	ctxx, span := r.Tracer.Start(ctxx, "ProfileRepository.GetBusinessUsers")

	collection := client.Database("PROFILE").Collection("user")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	filter := bson.D{{"role", "business"}}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		panic(err)
	}
	var users []model.UserBusiness
	if err = cursor.All(ctx, &users); err != nil {
		span.SetStatus(codes.Error, err.Error())
		fmt.Println(err)
	}
	span.SetStatus(codes.Ok, "")
	return users, nil
}

func (r ProfileRepository) GetBusinessUserByUsername(ctxx context.Context, username string) model.UserBusiness {
	ctxx, span := r.Tracer.Start(ctxx, "ProfileRepository.GetBusinessUserByUsername")

	collection := client.Database("PROFILE").Collection("user")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var user model.UserBusiness
	err := collection.FindOne(ctx, bson.D{{"username", username}}).Decode(&user)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		fmt.Println("Doesn't exist user with this username")
	}
	span.SetStatus(codes.Ok, "")
	return user
}

func (r ProfileRepository) GetBusinessUserById(ctxx context.Context, id string) model.UserBusiness {
	ctxx, span := r.Tracer.Start(ctxx, "ProfileRepository.GetBusinessUserById")

	collection := client.Database("PROFILE").Collection("user")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var user model.UserBusiness
	err := collection.FindOne(ctx, bson.D{{"id", id}}).Decode(&user)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		fmt.Println("Doesn't exist user with this id")
	}
	span.SetStatus(codes.Ok, "")
	return user
}

func (r ProfileRepository) SignUpBusinessUser(ctxx context.Context, user *model.UserBusiness) (*mongo.InsertOneResult, error) {
	ctxx, span := r.Tracer.Start(ctxx, "ProfileRepository.SignUpBusinessUser")

	collection := client.Database("PROFILE").Collection("user")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	user.PrivateProfile = true

	result, err := collection.InsertOne(ctx, user)
	span.SetStatus(codes.Ok, "")
	return result, err
}

//BOTH USER

func (r ProfileRepository) UpdateProfileStatus(ctxx context.Context, id string) error {
	ctxx, span := r.Tracer.Start(ctxx, "ProfileRepository.UpdateProfileStatus")

	collection := client.Database("PROFILE").Collection("user")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	user := r.GetRegularUserById(ctx, id)

	_, err := collection.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.D{
			{"$set", bson.D{{"privateProfile", !user.PrivateProfile}}},
		},
	)
	if err != nil {
		fmt.Println(err)
		span.SetStatus(codes.Error, err.Error())
	}
	fmt.Printf("Updated %v profile status!\n", user.Username)
	span.SetStatus(codes.Ok, "")
	return nil
}

func (r ProfileRepository) DeleteUser(ctxx context.Context, username string) {
	ctxx, span := r.Tracer.Start(ctxx, "ProfileRepository.DeleteUser")

	collection := client.Database("PROFILE").Collection("user")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	_, err := collection.DeleteOne(ctx, bson.M{"username": username})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		fmt.Println(err)
	}
	span.SetStatus(codes.Ok, "")
}

func (r ProfileRepository) DeleteUserById(ctxx context.Context, id string) {
	ctxx, span := r.Tracer.Start(ctxx, "ProfileRepository.DeleteUserById")

	collection := client.Database("PROFILE").Collection("user")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	_, err := collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		log.Fatal(err)
	}
	span.SetStatus(codes.Ok, "")
}

func (r ProfileRepository) GetListOfUsersByIds(ctxx context.Context, ids []string) ([]model.Users, error) {
	ctxx, span := r.Tracer.Start(ctxx, "ProfileRepository.GetListOfUsersByIds")

	var users []model.Users
	modelUser := model.Users{}

	for _, id := range ids {
		user := r.GetRegularUserById(ctxx, id)
		if user != (model.User{}) {
			if user.Role == "regular" {
				modelUser.Id = user.Id
				modelUser.Name = user.FirstName + " " + user.LastName
				modelUser.Username = user.Username
			}

			if user.Role == "business" {
				userBusiness := r.GetBusinessUserById(ctxx, id)
				modelUser.Id = userBusiness.Id
				modelUser.Name = userBusiness.Company
				modelUser.Username = userBusiness.Username
			}

			users = append(users, modelUser)
		}

	}
	span.SetStatus(codes.Ok, "")

	return users, nil
}

func (r ProfileRepository) GetUserById(ctxx context.Context, id string) model.Users {
	ctxx, span := r.Tracer.Start(ctxx, "ProfileRepository.GetUserById")

	modelUser := model.Users{}

	user := r.GetRegularUserById(ctxx, id)
	if user != (model.User{}) {
		if user.Role == "regular" {
			modelUser.Id = user.Id
			modelUser.Name = user.FirstName + " " + user.LastName
			modelUser.Username = user.Username
		}

		if user.Role == "business" {
			userBusiness := r.GetBusinessUserById(ctxx, id)
			modelUser.Id = userBusiness.Id
			modelUser.Name = userBusiness.Company
			modelUser.Username = userBusiness.Username
		}
	}
	span.SetStatus(codes.Ok, "")
	return modelUser
}

func Disconnect(ctx context.Context) error {
	err := client.Disconnect(ctx)
	if err != nil {
		return err
	}
	return nil
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
