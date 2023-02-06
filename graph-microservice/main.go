package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/ppetar33/twitter-api/graph-microservice/data"
	"github.com/ppetar33/twitter-api/graph-microservice/handlers"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func main() {

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

	tracer := tp.Tracer("graph-microservice")
	otel.SetTextMapPropagator(propagation.TraceContext{})

	graphRepository := data.GraphRepository{Tracer: tracer}
	graphHandler := handlers.GraphHandler{
		Tracer: tracer,
		Repo:   graphRepository,
	}

	fmt.Println(graphHandler)

	router := mux.NewRouter()
	router.Use(handlers.ExtractTraceInfoMiddleware)
	router.StrictSlash(true)

	router.HandleFunc("/api/socialGraph/registerUser", graphHandler.RegisterUser).Methods("POST")                                          // Register new entity
	router.Handle("/api/socialGraph/follow", ValidateJWT(graphHandler.Follow)).Methods("POST")                                             // Follow
	router.Handle("/api/socialGraph/unfollow", ValidateJWT(graphHandler.Unfollow)).Methods("POST")                                         // Unfollow
	router.Handle("/api/socialGraph/followRequest", ValidateJWT(graphHandler.RequestFollow)).Methods("POST")                               // Follow request
	router.Handle("/api/socialGraph/declineRequest", ValidateJWT(graphHandler.DeclineRequest)).Methods("POST")                             // Decline follow request
	router.Handle("/api/socialGraph/acceptRequest", ValidateJWT(graphHandler.AcceptRequest)).Methods("POST")                               // Accept followRequest
	router.Handle("/api/socialGraph/isFollowing/{requestedBy}/{requestedTo}", ValidateJWT(graphHandler.CheckRelathionship)).Methods("GET") // Status of two entities
	router.Handle("/api/socialGraph/getFollowing/{id}", ValidateJWT(graphHandler.GetAllFollowing)).Methods("GET")                          // Get entities that user follow
	router.Handle("/api/socialGraph/getFollowers/{id}", ValidateJWT(graphHandler.GetAllFollowers)).Methods("GET")                          // Get user followers
	router.Handle("/api/socialGraph/getRecommendations/{id}", ValidateJWT(graphHandler.GetAllRecommended)).Methods("GET")                  // Get recommended entities
	router.Handle("/api/socialGraph/getFollowRequests/{id}", ValidateJWT(graphHandler.GetAllRequests)).Methods("GET")                      // Get follow request
	router.HandleFunc("/api/socialGraph/deleteUser/{id}", graphHandler.DeleteUser).Methods("DELETE")                                       // Delete user

	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	startServer(router, cors)
}

func startServer(router *mux.Router, cors func(http.Handler) http.Handler) {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         "0.0.0.0:8082",
		Handler:      cors(router),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
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
			semconv.ServiceNameKey.String("graph-microservice"),
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
