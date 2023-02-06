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
	database "github.com/ppetar33/twitter-api/profile-microservice/server"
	server2 "github.com/ppetar33/twitter-api/profile-microservice/server"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func main() {
	log.Println("Starting the application")

	//file, errLog := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	//
	//if errLog != nil {
	//	log.Fatal(errLog)
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

	tracer := tp.Tracer("profile-microservice")
	otel.SetTextMapPropagator(propagation.TraceContext{})

	profileRepository := database.ProfileRepository{Tracer: tracer}
	profileHAndler := ProfileHAndler{
		Tracer: tracer,
		Repo:   profileRepository,
	}

	fmt.Println(profileHAndler)

	router := mux.NewRouter()
	router.Use(ExtractTraceInfoMiddleware)
	router.StrictSlash(true)

	server, err := server2.ConnectToProfileDatabase()

	if err != nil {
		log.Fatal(err)
	}

	log.Println(server.NumberSessionsInProgress())

	router.Handle("/api/profile/get-regular-users", ValidateJWT(profileHAndler.GetRegularUsers)).Methods("GET")
	router.Handle("/api/profile/get-business-users", ValidateJWT(profileHAndler.GetBusinessUsers)).Methods("GET")
	router.Handle("/api/profile/regular-user-username/{username}", ValidateJWT(profileHAndler.GetRegularUserByUsername)).Methods("GET")
	router.Handle("/api/profile/business-user-username/{username}", ValidateJWT(profileHAndler.GetBusinessUserByUsername)).Methods("GET")
	router.Handle("/api/profile/regular-user/{id}", ValidateJWT(profileHAndler.GetRegularUserByID)).Methods("GET")
	router.Handle("/api/profile/business-user/{id}", ValidateJWT(profileHAndler.GetBusinessUserById)).Methods("GET")
	router.HandleFunc("/api/profile/regular-user-sign-up", profileHAndler.SignUpRegularUser).Methods("POST")
	router.HandleFunc("/api/profile/business-user-sign-up", profileHAndler.SignUpBusinessUser).Methods("POST")
	router.Handle("/api/profile/update-profile-status/{id}", ValidateJWT(profileHAndler.UpdateProfileStatus)).Methods("PUT")
	router.Handle("/api/profile/get-users-by-ids", ValidateJWT(profileHAndler.GetListOfUsersByIds)).Methods("POST")
	router.Handle("/api/profile/get-user-by-id/{id}", ValidateJWT(profileHAndler.GetUserById)).Methods("GET")
	router.HandleFunc("/api/profile/cancel-user-register/{id}", profileHAndler.CancelUserRegister).Methods("DELETE")
	router.Handle("/api/profile/delete/{username}", ValidateJWT(profileHAndler.DeleteUser)).Methods("DELETE")

	cors := gorillaHandlers.CORS(gorillaHandlers.AllowedOrigins([]string{"*"}))

	startServer(router, cors)
}

func startServer(router *mux.Router, cors func(http.Handler) http.Handler) {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	srv := &http.Server{Addr: "0.0.0.0:8081", Handler: cors(router)}
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

	defer server2.Disconnect(ctx)

	// NoSQL: Checking if the connection was established
	server2.Ping()

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
			semconv.ServiceNameKey.String("profile-microservice"),
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
