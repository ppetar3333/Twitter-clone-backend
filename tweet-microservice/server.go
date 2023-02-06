package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/gorilla/mux"
	model2 "github.com/ppetar33/twitter-api/tweet-microservice/model"
	database "github.com/ppetar33/twitter-api/tweet-microservice/server"
)

type TweetHandler struct {
	Tracer trace.Tracer
	Repo   database.TweetRepository
}

const contentType = "Content Type"
const applicationJson = "application/json"
const userErrorMessage = "User id is missing!"
const missingId = "ID is missing in parameters"
const contentTypeGet = "Content-Type"

func ValidateRequest(response http.ResponseWriter, request *http.Request) bool {
	response.Header().Set(contentType, applicationJson)
	contentType := request.Header.Get(contentTypeGet)
	mediatype, _, errContentType := mime.ParseMediaType(contentType)

	if errContentType != nil {
		http.Error(response, errContentType.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return true
	}

	if mediatype != applicationJson {
		err := errors.New("expect application/json Content-Type")
		http.Error(response, err.Error(), http.StatusUnsupportedMediaType)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusUnsupportedMediaType))
		return true
	}
	return false
}

func (h TweetHandler) GetAllTweets(response http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "TweetHandler.GetAllTweets")

	tweets, err := h.Repo.GetAllTweets(ctx)
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, err.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	} else {
		if tweets == nil {
			span.SetStatus(codes.Ok, "")
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
			json.NewEncoder(response).Encode(`[]`)
		} else {
			span.SetStatus(codes.Ok, "")
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
			json.NewEncoder(response).Encode(tweets)
		}
	}
}

func (h TweetHandler) CreateTweet(response http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "TweetHandler.CreateTweet")

	if ValidateRequest(response, request) {
		return
	}

	tweet, errDecodeBody := DecodeBody(request.Body)

	if errDecodeBody != nil {
		http.Error(response, errDecodeBody.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errDecodeBody.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	var textCheck = tweet.Text
	strings.ReplaceAll(textCheck, " ", "")

	if tweet != nil && textCheck != "" {
		result, err := h.Repo.CreateTweet(ctx, tweet)

		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			span.SetStatus(codes.Error, err.Error())
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
			return
		}
		span.SetStatus(codes.Ok, "")
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
		json.NewEncoder(response).Encode(result)
	} else {
		http.Error(response, "", http.StatusBadRequest)
		return
	}
}

func (h TweetHandler) GetAllTweetsForUser(response http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "TweetHandler.GetAllTweetsForUser")

	response.Header().Set(contentTypeGet, applicationJson)

	vars := mux.Vars(request)
	userid, userIdOk := vars["userid"]
	if !userIdOk {
		json.NewEncoder(response).Encode(`User id is missing!`)
		span.SetStatus(codes.Ok, userErrorMessage)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	usersFollowing, errFollowingUsers := h.Repo.GetAllFollowing(ctx, userid, request.Header["Token"][0])

	fmt.Println("USERS FOLLOWINGS ", usersFollowing)

	if errFollowingUsers != nil {
		http.Error(response, errFollowingUsers.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errFollowingUsers.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	userFounded, userFoundedErr := h.Repo.GetUserById(ctx, userid, request.Header["Token"][0])

	if userFoundedErr != nil {
		http.Error(response, userFoundedErr.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, userFoundedErr.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	usersFollowing = append(usersFollowing, userFounded)

	usersFollowingTweets, errUsersFollowingTweets := h.Repo.GetTweetsByUsersIds(ctx, usersFollowing, userid)

	if errUsersFollowingTweets != nil {
		http.Error(response, errUsersFollowingTweets.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errUsersFollowingTweets.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	tweetsWithUser, errorTweetsWithUser := h.Repo.GetTweetsWithUser(ctx, usersFollowingTweets, request.Header["Token"][0])

	if errorTweetsWithUser != nil {
		http.Error(response, errorTweetsWithUser.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errorTweetsWithUser.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	tweetsWithRetweets, errorTweetsWithRetweets := h.Repo.GetRetweetsForTweet(ctx, tweetsWithUser, request.Header["Token"][0])

	fmt.Println("TWEETS WITH RETWEETS ", tweetsWithRetweets)

	if errorTweetsWithRetweets != nil {
		http.Error(response, errorTweetsWithRetweets.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errorTweetsWithRetweets.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	tweetsForFeed, errorTweetsForFeed := h.Repo.GetTweetsForFeed(ctx, usersFollowing, tweetsWithRetweets)

	fmt.Println("TWEETS FOR FEED ", tweetsForFeed)

	if errorTweetsForFeed != nil {
		http.Error(response, errorTweetsForFeed.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errorTweetsForFeed.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	sort.Slice(tweetsForFeed, func(i, j int) bool {
		return tweetsForFeed[j].CreatedAt.Before(tweetsForFeed[i].CreatedAt)
	})

	log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
	json.NewEncoder(response).Encode(tweetsForFeed)
	span.SetStatus(codes.Ok, "")
}

func (h TweetHandler) GetAllTweetsForProfilePage(response http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "TweetHandler.GetAllTweetsForProfilePage")

	response.Header().Set(contentTypeGet, applicationJson)

	vars := mux.Vars(request)
	userid, userIdOk := vars["userprofileid"]
	loggedInUserId, loggedInUserIdOK := vars["userid"]

	if !userIdOk || !loggedInUserIdOK {
		json.NewEncoder(response).Encode(`User id is missing!`)
		span.SetStatus(codes.Ok, userErrorMessage)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if loggedInUserId == userid {
		h.GetAllTweetsForLoggedInUser(response, request)
		return
	}

	usersFollowing, errFollowingUsers := h.Repo.GetAllFollowing(ctx, loggedInUserId, request.Header["Token"][0])

	fmt.Println("USERS FOLLOWINGS ", usersFollowing)

	if errFollowingUsers != nil {
		http.Error(response, errFollowingUsers.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errFollowingUsers.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	var usersList []model2.UserFollowing

	userFounded, userFoundedErr := h.Repo.GetUserById(ctx, userid, request.Header["Token"][0])

	if userFoundedErr != nil {
		http.Error(response, userFoundedErr.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errFollowingUsers.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	usersList = append(usersList, userFounded)

	usersFollowingTweets, errUsersFollowingTweets := h.Repo.GetTweetsByUsersIds(ctx, usersList, userid)

	if errUsersFollowingTweets != nil {
		http.Error(response, errUsersFollowingTweets.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errFollowingUsers.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	tweetsWithUser, errorTweetsWithUser := h.Repo.GetTweetsWithUser(ctx, usersFollowingTweets, request.Header["Token"][0])

	if errorTweetsWithUser != nil {
		http.Error(response, errorTweetsWithUser.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errFollowingUsers.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	tweetsWithRetweets, errorTweetsWithRetweets := h.Repo.GetRetweetsForTweet(ctx, tweetsWithUser, request.Header["Token"][0])

	if errorTweetsWithRetweets != nil {
		http.Error(response, errorTweetsWithRetweets.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errFollowingUsers.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	tweetsForUserProfile, errorTweetsForProfile := h.Repo.GetTweetsForFeed(ctx, usersFollowing, tweetsWithRetweets)

	fmt.Println("TWEETS FOR PROFILE ", tweetsForUserProfile)

	if errorTweetsForProfile != nil {
		http.Error(response, errorTweetsForProfile.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errFollowingUsers.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	sort.Slice(tweetsForUserProfile, func(i, j int) bool {
		return tweetsForUserProfile[j].CreatedAt.Before(tweetsForUserProfile[i].CreatedAt)
	})

	log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
	json.NewEncoder(response).Encode(tweetsForUserProfile)
	span.SetStatus(codes.Ok, "")

}

func (h TweetHandler) GetTweetById(response http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "TweetHandler.GetTweetById")
	response.Header().Set(contentTypeGet, applicationJson)

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		json.NewEncoder(response).Encode(`ID is missing in parameters`)
		span.SetStatus(codes.Ok, missingId)
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	tweet, err := h.Repo.GetById(ctx, id)
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, err.Error())
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}
	if tweet == (model2.Tweet{}) {
		json.NewEncoder(response).Encode(`Doesn't exist tweet with this id`)
		span.SetStatus(codes.Ok, "Doesn't exist tweet with this id")
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	} else {
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusOK))
		span.SetStatus(codes.Ok, "")
		json.NewEncoder(response).Encode(tweet)
	}
}

func (h TweetHandler) GetAllTweetsForLoggedInUser(response http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "TweetHandler.GetAllTweetsForProfilePage")

	response.Header().Set(contentTypeGet, applicationJson)

	vars := mux.Vars(request)
	userid, userIdOk := vars["userid"]
	if !userIdOk {
		json.NewEncoder(response).Encode(`User id is missing!`)
		span.SetStatus(codes.Ok, userErrorMessage)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	usersFollowing := []model2.UserFollowing{}

	userFounded, userFoundedErr := h.Repo.GetUserById(ctx, userid, request.Header["Token"][0])

	if userFoundedErr != nil {
		http.Error(response, userFoundedErr.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, userFoundedErr.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	usersFollowing = append(usersFollowing, userFounded)

	usersFollowingTweets, errUsersFollowingTweets := h.Repo.GetTweetsByUsersIds(ctx, usersFollowing, userid)

	if errUsersFollowingTweets != nil {
		http.Error(response, errUsersFollowingTweets.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errUsersFollowingTweets.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	tweetsWithUser, errorTweetsWithUser := h.Repo.GetTweetsWithUser(ctx, usersFollowingTweets, request.Header["Token"][0])

	if errorTweetsWithUser != nil {
		http.Error(response, errorTweetsWithUser.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errorTweetsWithUser.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	tweetsWithRetweets, errorTweetsWithRetweets := h.Repo.GetRetweetsForTweet(ctx, tweetsWithUser, request.Header["Token"][0])

	if errorTweetsWithRetweets != nil {
		http.Error(response, errorTweetsWithRetweets.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errorTweetsWithRetweets.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	followings, errFollowing := h.Repo.GetAllFollowing(ctx, userid, request.Header["Token"][0])

	if errFollowing != nil {
		http.Error(response, errFollowing.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errFollowing.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	tweetsForFeed, errorTweetsForFeed := h.Repo.GetTweetsForFeed(ctx, followings, tweetsWithRetweets)

	fmt.Println("TWEETS FOR FEED ", tweetsForFeed)

	if errorTweetsForFeed != nil {
		http.Error(response, errorTweetsForFeed.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errorTweetsForFeed.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	sort.Slice(tweetsForFeed, func(i, j int) bool {
		return tweetsForFeed[j].CreatedAt.Before(tweetsForFeed[i].CreatedAt)
	})

	log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
	json.NewEncoder(response).Encode(tweetsForFeed)
	span.SetStatus(codes.Ok, "")
}

func (h TweetHandler) LikeUnlikeTweet(response http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "TweetHandler.LikeUnlikeTweet")

	response.Header().Set(contentTypeGet, applicationJson)

	vars := mux.Vars(r)
	tweetId, ok := vars["tweetId"]
	if !ok {
		json.NewEncoder(response).Encode(`Tweet ID is missing in parameters`)
		span.SetStatus(codes.Ok, missingId)
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	userId, ok1 := vars["userId"]
	if !ok1 {
		json.NewEncoder(response).Encode(`User ID is missing in parameters`)
		span.SetStatus(codes.Ok, missingId)
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	message, err := h.Repo.LikeUnlikeTweet(ctx, tweetId, userId)
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, err.Error())
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	} else {
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusOK))
		span.SetStatus(codes.Ok, "")
		json.NewEncoder(response).Encode(message)
	}
}

func ExtractTraceInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
