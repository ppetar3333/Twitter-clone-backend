package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"github.com/ppetar33/twitter-api/tweet-microservice/model"
	"github.com/sony/gobreaker"
)

type DBConnection struct {
	session *gocql.Session
	cluster *gocql.ClusterConfig
}

var connection DBConnection
var cb = CircuitBreaker()

type TweetRepository struct {
	Tracer trace.Tracer
}

const errorReadingBody = "Error reading body"
const errorUnmarshalingBody = "Error unmarshaling body"

func SetupDBConnection() {
	connection.cluster = gocql.NewCluster("tweet-db:9042")
	connection.cluster.ProtoVersion = 4
	connection.cluster.DisableInitialHostLookup = true
	connection.cluster.Consistency = gocql.Quorum
	connection.cluster.Keyspace = "twitter"

	connection.session, _ = connection.cluster.CreateSession()

	/*
		Ako sistem ne radi uopšte kada se na prvu pokrene uraditi sledeće:
			1. Otići u docker desktop
			2. Ući u tweet-db kontejner i ući u terminal
			3. Komandom cqlsh aktivirati cassandra terminal
			4. Uraditi describe keyspaces
			5. Ako ne postoji keyspace twitter u navodima uraditi sledeću kompandu
			6. CREATE KEYSPACE IF NOT EXISTS twitter WITH replication = {'class': 'SimpleStrategy', 'replication_factor' : 1}; (bez prefiksa cqlsh jer je terminal već aktiviran
			7. Za sada znamo da radi jer ne vraća više 502 grešku
	*/
	err := connection.session.Query("CREATE KEYSPACE IF NOT EXISTS twitter WITH replication = {'class': 'SimpleStrategy', 'replication_factor' : 1};").Exec()
	if err != nil {
		fmt.Println(err)
	}

	err = connection.session.Query("CREATE TABLE IF NOT EXISTS twitter.tweets ( id UUID PRIMARY KEY, tweetText text, user text, tweet text, likes list<text>, createdAt timestamp)").Exec()
	if err != nil {
		fmt.Println(err)
	}
}

func ExecuteQuery(query string, values ...interface{}) error {
	if err := connection.session.Query(query).Bind(values...).Exec(); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (r TweetRepository) CreateTweet(ctx context.Context, tweet *model.Tweet) (string, error) {
	ctx, span := r.Tracer.Start(ctx, "TweetRepository.CreateTweet")

	query := `INSERT INTO tweets( id, tweetText, user, tweet, likes, createdAt) VALUES (?, ?, ?, ?, ?, ?)`

	//t, _ := time.Parse("1999-12-31", tweet.CreatedAt)
	//fmt.Println(t)
	err := ExecuteQuery(query, uuid.New().String(), tweet.Text, tweet.User, tweet.Tweet, tweet.Likes, time.Now())

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", err
	}
	span.SetStatus(codes.Ok, "")
	return "OK", nil
}

func (r TweetRepository) GetAllTweets(ctx context.Context) ([]model.Tweet, error) {
	ctx, span := r.Tracer.Start(ctx, "TweetRepository.GetAllTweets")
	localCql := "SELECT id, createdat, likes, tweettext, tweet, user FROM tweets;"
	var tweet model.Tweet
	var tweets []model.Tweet

	iter := connection.session.Query(localCql).Iter()

	for iter.Scan(&tweet.Id, &tweet.CreatedAt, &tweet.Likes, &tweet.Text, &tweet.Tweet, &tweet.User) {
		tweets = append(tweets, tweet)
	}

	if err := iter.Close(); err != nil {
		span.SetStatus(codes.Error, err.Error())
		fmt.Println(err)
	}

	fmt.Println(tweets)
	span.SetStatus(codes.Ok, "")
	return tweets, nil
}

func (r TweetRepository) GetAllFollowing(ctx context.Context, userid string, token string) ([]model.UserFollowing, error) {
	ctx, span := r.Tracer.Start(ctx, "TweetRepository.GetAllFollowing")

	client := http.Client{
		Timeout: time.Second * 60000,
	}

	responseGraph, errGraph := cb.Execute(func() (interface{}, error) {
		path := "http://graph-microservice:8082/api/socialGraph/getFollowing/" + userid
		req, _ := http.NewRequest("GET", path, nil)
		req.Header.Set("Token", token)
		resp, err := client.Do(req)

		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("error status: " + resp.Status)
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return bodyBytes, nil
	})

	//path := "http://graph-microservice:8082/api/socialGraph/getFollowing/" + userid
	//responseGraph, errGraph := http.Get(path)

	if errGraph != nil {
		span.SetStatus(codes.Error, errGraph.Error())
		fmt.Println("Error with response from graph service", errGraph)
	}

	//defer responseGraph.Body.Close()
	reader := bytes.NewReader(responseGraph.([]byte))
	body, errBody := ioutil.ReadAll(reader)

	if errBody != nil {
		span.SetStatus(codes.Error, errGraph.Error())
		fmt.Println(errorReadingBody, errBody)
	}
	var usersFollowing []model.UserFollowing

	errUnmarshal := json.Unmarshal(body, &usersFollowing)

	if errUnmarshal != nil {
		span.SetStatus(codes.Error, errGraph.Error())
		fmt.Println(errorUnmarshalingBody, errUnmarshal)
	}

	span.SetStatus(codes.Ok, "")
	return usersFollowing, nil
}

func (r TweetRepository) GetUserById(ctx context.Context, userid string, token string) (model.UserFollowing, error) {
	ctx, span := r.Tracer.Start(ctx, "TweetRepository.GetUserById")

	client := http.Client{
		Timeout: time.Second * 60000,
	}
	responseProfile, errProfile := cb.Execute(func() (interface{}, error) {
		path := "http://profile-microservice:8081/api/profile/get-user-by-id/" + userid
		req, _ := http.NewRequest("GET", path, nil)
		req.Header.Set("Token", token)
		resp, err := client.Do(req)

		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("error status: " + resp.Status)
		}
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return bodyBytes, nil
	})

	//path := "http://profile-microservice:8081/api/profile/get-user-by-id/" + userid
	//responseProfile, errProfile := http.Get(path)

	if errProfile != nil {
		span.SetStatus(codes.Error, errProfile.Error())
		fmt.Println("Error with response from profile service", errProfile)
	}

	reader := bytes.NewReader(responseProfile.([]byte))
	body, errBody := ioutil.ReadAll(reader)

	if errBody != nil {
		span.SetStatus(codes.Error, errBody.Error())
		fmt.Println(errorReadingBody, errBody)
	}

	var userProfile model.UserProfile

	errUnmarshal := json.Unmarshal(body, &userProfile)

	if errUnmarshal != nil {
		span.SetStatus(codes.Error, errUnmarshal.Error())
		fmt.Println(errorUnmarshalingBody, errUnmarshal)
	}

	var userReturn model.UserFollowing

	userReturn.Id = userProfile.Id
	userReturn.Name = userProfile.Name
	userReturn.Username = userProfile.Username
	userReturn.Type = ""
	span.SetStatus(codes.Ok, "")
	return userReturn, nil
}

func (r TweetRepository) GetTweetsByUsersIds(ctx context.Context, users []model.UserFollowing, loggedInUser string) ([]model.Tweet, error) {
	ctx, span := r.Tracer.Start(ctx, "TweetRepository.GetTweetsByUsersIds")

	var tweets []model.Tweet

	tweetsDatabase, errTweets := r.GetAllTweets(ctx)

	if errTweets != nil {
		span.SetStatus(codes.Error, errTweets.Error())
		fmt.Println("Error getting all of the tweets", errTweets)
	}

	for _, user := range users {
		for _, tweet := range tweetsDatabase {
			if user.Id == tweet.User {
				tweets = append(tweets, tweet)
			}
		}
	}
	span.SetStatus(codes.Ok, "")
	return tweets, nil
}

func (r TweetRepository) GetTweetsWithUser(ctx context.Context, tweets []model.Tweet, token string) ([]model.TweetWithUser, error) {
	ctx, span := r.Tracer.Start(ctx, "TweetRepository.GetTweetsWithUser")

	client := http.Client{
		Timeout: time.Second * 60000,
	}
	var tweetsReturn []model.TweetWithUser

	for _, tweet := range tweets {

		path := "http://profile-microservice:8081/api/profile/get-user-by-id/" + tweet.User
		req, _ := http.NewRequest("GET", path, nil)
		req.Header.Set("Token", token)
		responseProfile, errProfile := client.Do(req)

		if errProfile != nil {
			fmt.Println("Error with response from profile service", errProfile)
		}

		body, errBody := ioutil.ReadAll(responseProfile.Body)

		if errBody != nil {
			span.SetStatus(codes.Error, errBody.Error())
			fmt.Println(errorReadingBody, errBody)
		}

		var userProfile model.UserProfile

		errUnmarshal := json.Unmarshal(body, &userProfile)

		if errUnmarshal != nil {
			span.SetStatus(codes.Error, errUnmarshal.Error())
			fmt.Println(errorUnmarshalingBody, errUnmarshal)
		}

		var tweetReturn model.TweetWithUser

		tweetReturn.Id = tweet.Id
		tweetReturn.Text = tweet.Text
		tweetReturn.Likes = tweet.Likes
		tweetReturn.Tweet = tweet.Tweet
		tweetReturn.User = userProfile
		tweetReturn.CreatedAt = tweet.CreatedAt

		tweetsReturn = append(tweetsReturn, tweetReturn)
	}
	span.SetStatus(codes.Ok, "")
	return tweetsReturn, nil
}

func (r TweetRepository) GetRetweetsForTweet(ctx context.Context, tweets []model.TweetWithUser, token string) ([]model.TweetWithRetweet, error) {
	ctx, span := r.Tracer.Start(ctx, "TweetRepository.GetRetweetsForTweet")

	var tweetsReturn []model.TweetWithRetweet

	for _, tweet := range tweets {
		if tweet.Tweet != "" { // if retweet exists
			tweetFounded, errFoundingTweet := r.GetById(ctx, tweet.Tweet)

			fmt.Println("RETWEET FOUNDED", tweetFounded)

			if errFoundingTweet != nil {
				span.SetStatus(codes.Error, errFoundingTweet.Error())
				fmt.Println("Error getting tweet by id", errFoundingTweet)
			}

			var tweetReturn model.TweetWithRetweet

			var tweetFoundedList []model.Tweet
			tweetFoundedList = append(tweetFoundedList, tweetFounded)

			fmt.Println("RETWEET FOUNDED LIST", tweetFoundedList)

			tweetWithUser, errTweetWithUser := r.GetTweetsWithUser(ctx, tweetFoundedList, token)

			fmt.Println("RETWEET WITH USER", tweetWithUser)

			if errTweetWithUser != nil {
				span.SetStatus(codes.Error, errTweetWithUser.Error())
				fmt.Println("Error getting tweet with user", errTweetWithUser)
			}

			tweetReturn.Id = tweet.Id
			tweetReturn.Text = tweet.Text
			tweetReturn.Likes = tweet.Likes
			tweetReturn.Tweet = tweetWithUser[0]
			tweetReturn.User = tweet.User
			tweetReturn.CreatedAt = tweet.CreatedAt

			tweetsReturn = append(tweetsReturn, tweetReturn)
		} else {
			var tweetReturn model.TweetWithRetweet

			tweetReturn.Id = tweet.Id
			tweetReturn.Text = tweet.Text
			tweetReturn.Likes = tweet.Likes
			tweetReturn.Tweet = model.TweetWithUser{}
			tweetReturn.User = tweet.User
			tweetReturn.CreatedAt = tweet.CreatedAt

			tweetsReturn = append(tweetsReturn, tweetReturn)
		}
	}
	span.SetStatus(codes.Ok, "")
	return tweetsReturn, nil
}

func (r TweetRepository) GetTweetsForFeed(ctx context.Context, users []model.UserFollowing, tweets []model.TweetWithRetweet) ([]model.TweetWithRetweet, error) {
	ctx, span := r.Tracer.Start(ctx, "TweetRepository.GetTweetsForFeed")

	var tweetsReturn []model.TweetWithRetweet

	var private []model.TweetWithRetweet
	var public []model.TweetWithRetweet

	for _, user := range users {
		for _, tweet := range tweets {
			tweetReduced := tweet

			if user.Id != tweet.Tweet.User.Id {
				tweetReduced.Tweet.Id = ""
				tweetReduced.Tweet.Text = ""
				var likes *[]string
				tweetReduced.Tweet.Likes = likes
				private = append(private, tweetReduced)
			} else {
				public = append(public, tweetReduced)
			}

		}
	}

	tweetsReturn = append(private, public...)

	fmt.Println("PUBLIC TWEETS: ", public)
	fmt.Println("PRIVATE TWEETS: ", private)

	result := RemoveDuplicated(tweetsReturn)
	span.SetStatus(codes.Ok, "")
	return result, nil
}

func RemoveDuplicated(sample []model.TweetWithRetweet) []model.TweetWithRetweet {
	var unique []model.TweetWithRetweet
sampleLoop:
	for _, v := range sample {
		for i, u := range unique {
			if v.Id == u.Id {
				unique[i] = v
				continue sampleLoop
			}
		}
		unique = append(unique, v)
	}
	return unique
}

func (r TweetRepository) GetById(ctx context.Context, id string) (model.Tweet, error) {
	ctx, span := r.Tracer.Start(ctx, "TweetRepository.GetById")

	var tweet model.Tweet

	iter := connection.session.Query(`SELECT id, createdat, likes, tweettext, tweet, user FROM tweets WHERE id=?`, id).Iter()

	for iter.Scan(&tweet.Id, &tweet.CreatedAt, &tweet.Likes, &tweet.Text, &tweet.Tweet, &tweet.User) {
		fmt.Println("Tweet: ", tweet)
	}
	span.SetStatus(codes.Ok, "")
	return tweet, nil
}

func (r TweetRepository) LikeUnlikeTweet(ctx context.Context, tweetId string, userId string) (string, error) {
	ctx, span := r.Tracer.Start(ctx, "TweetRepository.LikeUnlikeTweet")

	tweet, err := r.GetById(ctx, tweetId)
	if err != nil {
		fmt.Println("Error getting tweet by id", err)
		span.SetStatus(codes.Error, err.Error())
	}

	alreadyLike := false
	message := ""

	if tweet.Likes == nil {

		var a = []string{userId}
		tweet.Likes = &a
		if err := connection.session.Query(`UPDATE tweets SET likes=? WHERE id=?`, tweet.Likes, tweetId).Exec(); err != nil {
			fmt.Println(err)
		}
		message = "SUCCESSFUL LIKE POST"
	} else {
		for _, id := range *tweet.Likes {
			if id == userId {
				alreadyLike = true
			}
		}

		if alreadyLike == false {
			*tweet.Likes = append(*tweet.Likes, userId)
			if err := connection.session.Query(`UPDATE tweets SET likes=? WHERE id=?`, tweet.Likes, tweetId).Exec(); err != nil {
				fmt.Println(err)
			}
			message = "SUCCESSFUL LIKE POST"
		} else if alreadyLike == true {

			for i, v := range *tweet.Likes {
				if v == userId {
					*tweet.Likes = append((*tweet.Likes)[:i], (*tweet.Likes)[i+1:]...)
					break
				}
			}

			if err := connection.session.Query(`UPDATE tweets SET likes=? WHERE id=?`, tweet.Likes, tweetId).Exec(); err != nil {
				fmt.Println(err)
			}
			message = "SUCCESSFUL UNLIKE POST"
		}
	}
	span.SetStatus(codes.Ok, "")
	return message, nil
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
				fmt.Println("Circuit Breaker '%s' changed from '%s' to '%s'\n", name, from, to)
			},
		},
	)
}
