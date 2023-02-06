package main

import (
	"context"
	"fmt"
	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	database "github.com/ppetar33/twitter-api/tweet-microservice/server"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("Starting the application")

	//file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	//
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.SetOutput(file)

	ctx := context.Background()
	exp, errExporter := newExporter("http://jaeger:14268/api/traces")
	if errExporter != nil {
		fmt.Println("failed to initialize exporter: %v", errExporter)
	}

	tp := newTraceProvider(exp)

	defer func() { _ = tp.Shutdown(ctx) }()
	otel.SetTracerProvider(tp)

	tracer := tp.Tracer("tweet-microservice")
	otel.SetTextMapPropagator(propagation.TraceContext{})

	authRepository := database.TweetRepository{Tracer: tracer}
	authHandler := TweetHandler{
		Tracer: tracer,
		Repo:   authRepository,
	}

	fmt.Println(authHandler)

	router := mux.NewRouter()
	router.Use(ExtractTraceInfoMiddleware)
	router.StrictSlash(true)

	database.SetupDBConnection()

	router.Handle("/api/tweets/get-all", ValidateJWT(authHandler.GetAllTweets)).Methods("GET")
	router.Handle("/api/tweets/get-all/{userid}", ValidateJWT(authHandler.GetAllTweetsForUser)).Methods("GET")
	router.Handle("/api/tweets/get-all-of-user/{userprofileid}/{userid}", ValidateJWT(authHandler.GetAllTweetsForProfilePage)).Methods("GET")
	router.Handle("/api/tweets/create", ValidateJWT(authHandler.CreateTweet)).Methods("POST")
	router.Handle("/api/tweets/get-tweet-by-id/{id}", ValidateJWT(authHandler.GetTweetById)).Methods("GET")
	router.Handle("/api/tweets/like-unlike-tweet/{tweetId}/{userId}", ValidateJWT(authHandler.LikeUnlikeTweet)).Methods("PUT")

	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	startServer(router, cors)
}

func startServer(router *mux.Router, cors func(http.Handler) http.Handler) {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         "0.0.0.0:8083",
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second, // from 1 to 5 changed
		WriteTimeout: 5 * time.Second, // from 1 to 5 chenged
	}

	go func() {
		log.Println("Server Starting")
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}
	}()

	shutDownServer(srv, quit)
}

func shutDownServer(srv *http.Server, quit chan os.Signal) {
	<-quit

	log.Println("Service Shutting Down ...")

	// gracefully stop server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server Stopped")
}

func newExporter(address string) (*jaeger.Exporter, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(address)))
	if err != nil {
		return nil, err
	}
	return exp, nil
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("tweet-microservice"),
		),
	)

	if err != nil {
		panic(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)
}
