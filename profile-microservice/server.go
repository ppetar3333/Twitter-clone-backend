package main

import (
	"encoding/json"
	"errors"
	"log"
	"mime"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/ppetar33/twitter-api/profile-microservice/model"
	database "github.com/ppetar33/twitter-api/profile-microservice/server"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type ProfileHAndler struct {
	Tracer trace.Tracer
	Repo   database.ProfileRepository
}

const applicationJson = "application/json"
const contentType = "Content-Type"
const expectJson = "expect application/json Content-Type"

// REGULAR USER
func (h ProfileHAndler) GetRegularUsers(response http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "ProfileHAndler.GetRegularUsers")

	response.Header().Set(contentType, applicationJson)

	users, err := h.Repo.GetRegularUsers(ctx)

	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		span.SetStatus(codes.Error, err.Error())
		return
	} else {
		if users == nil {
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
			json.NewEncoder(response).Encode(`[]`)
			span.SetStatus(codes.Ok, "")
		} else {
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
			json.NewEncoder(response).Encode(users)
			span.SetStatus(codes.Ok, "")
		}
	}
}
func (h ProfileHAndler) GetRegularUserByUsername(response http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "ProfileHAndler.GetRegularUserByUsername")

	response.Header().Set(contentType, applicationJson)

	vars := mux.Vars(r)
	username, ok := vars["username"]
	if !ok {
		json.NewEncoder(response).Encode(`Username is missing in parameters`)
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	user := h.Repo.GetRegularUserByUsername(ctx, username)

	if user == (model.User{}) {
		json.NewEncoder(response).Encode(`Doesn't exist user with this username`)
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	} else {
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusOK))
		span.SetStatus(codes.Ok, "")
		json.NewEncoder(response).Encode(user)
	}
}

func (h ProfileHAndler) GetRegularUserByID(response http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "ProfileHAndler.GetRegularUserByID")

	response.Header().Set(contentType, applicationJson)

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		json.NewEncoder(response).Encode(`ID is missing in parameters`)
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	user := h.Repo.GetRegularUserById(ctx, id)
	if user == (model.User{}) {
		json.NewEncoder(response).Encode(`Doesn't exist user with this id`)
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	} else {
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusOK))
		span.SetStatus(codes.Ok, "")
		json.NewEncoder(response).Encode(user)
	}
}

func (h ProfileHAndler) SignUpRegularUser(response http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "ProfileHAndler.SignUpRegularUser")

	response.Header().Set(contentType, applicationJson)
	contentType := r.Header.Get(contentType)
	mediatype, _, errContentType := mime.ParseMediaType(contentType)

	if errContentType != nil {
		http.Error(response, errContentType.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errContentType.Error())
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if mediatype != applicationJson {
		err := errors.New(expectJson)
		http.Error(response, err.Error(), http.StatusUnsupportedMediaType)
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusUnsupportedMediaType))
		return
	}

	user, errDecodeBody := DecodeBodyRegister(r.Body)

	if errDecodeBody != nil {
		http.Error(response, errDecodeBody.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errDecodeBody.Error())
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if user != nil {

		errMessage := SignUpValidationRegularUser(user)
		if errMessage != "" {
			json.NewEncoder(response).Encode(errMessage)
			span.SetStatus(codes.Error, errMessage)
			log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
			return
		} else {
			_, err := h.Repo.SignUpRegularUser(ctx, user)
			if err != nil {
				json.NewEncoder(response).Encode(err)
				span.SetStatus(codes.Error, err.Error())
				log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
				return
			} else {
				log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusOK))
				json.NewEncoder(response).Encode(`Successful sign up!`)
				span.SetStatus(codes.Ok, "")
			}
		}
	}
}

//BUSINESS USER

func (h ProfileHAndler) GetBusinessUsers(response http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "ProfileHAndler.GetBusinessUsers")

	response.Header().Set(contentType, applicationJson)

	users, err := h.Repo.GetBusinessUsers(ctx)

	if err != nil {
		json.NewEncoder(response).Encode(err)
		span.SetStatus(codes.Error, err.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	} else {
		if users == nil {
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
			json.NewEncoder(response).Encode(`[]`)
			span.SetStatus(codes.Ok, "")
		} else {
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
			json.NewEncoder(response).Encode(users)
			span.SetStatus(codes.Ok, "")
		}
	}
}

func (h ProfileHAndler) GetBusinessUserByUsername(response http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "ProfileHAndler.GetBusinessUserByUsername")

	response.Header().Set(contentType, applicationJson)

	vars := mux.Vars(r)
	username, ok := vars["username"]
	if !ok {
		json.NewEncoder(response).Encode(`Username is missing in parameters`)
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	user := h.Repo.GetBusinessUserByUsername(ctx, username)
	if user == (model.UserBusiness{}) {
		json.NewEncoder(response).Encode(`Doesn't exist user with this username`)
		span.SetStatus(codes.Ok, "")
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	} else {
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusOK))
		json.NewEncoder(response).Encode(user)
		span.SetStatus(codes.Ok, "")
	}
}

func (h ProfileHAndler) GetBusinessUserById(response http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "ProfileHAndler.GetBusinessUserById")

	response.Header().Set(contentType, applicationJson)

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		json.NewEncoder(response).Encode(`ID is missing in parameters`)
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	user := h.Repo.GetBusinessUserById(ctx, id)

	if user == (model.UserBusiness{}) {
		json.NewEncoder(response).Encode(`Doesn't exist user with this id`)
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	} else {
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusOK))
		json.NewEncoder(response).Encode(user)
		span.SetStatus(codes.Ok, "")
	}
}

func (h ProfileHAndler) SignUpBusinessUser(response http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "ProfileHAndler.SignUpBusinessUser")

	response.Header().Set(contentType, applicationJson)
	contentType := r.Header.Get(contentType)
	mediatype, _, errContentType := mime.ParseMediaType(contentType)

	if errContentType != nil {
		http.Error(response, errContentType.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errContentType.Error())
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if mediatype != applicationJson {
		err := errors.New(expectJson)
		http.Error(response, err.Error(), http.StatusUnsupportedMediaType)
		span.SetStatus(codes.Error, err.Error())
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusUnsupportedMediaType))
		return
	}

	user, errDecodeBody := DecodeBodyRegisterBusinessUser(r.Body)

	if errDecodeBody != nil {
		http.Error(response, errDecodeBody.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errDecodeBody.Error())
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if user != nil {
		errMessage := SignUpValidationBusinessUser(user)
		if errMessage != "" {
			json.NewEncoder(response).Encode(errMessage)
			span.SetStatus(codes.Error, errMessage)
			log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
			return
		} else {
			_, err := h.Repo.SignUpBusinessUser(ctx, user)
			if err != nil {
				json.NewEncoder(response).Encode(err)
				span.SetStatus(codes.Error, err.Error())
				log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
				return
			} else {
				log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusOK))
				json.NewEncoder(response).Encode(`Successful sign up!`)
				span.SetStatus(codes.Ok, "")
			}
		}
	}
}

// BOTH USER
func (h ProfileHAndler) UpdateProfileStatus(response http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "ProfileHAndler.UpdateProfileStatus")

	response.Header().Set(contentType, applicationJson)

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		json.NewEncoder(response).Encode(`*** ID is missing in parameters ***`)
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	err := h.Repo.UpdateProfileStatus(ctx, id)
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, err.Error())
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	} else {
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusOK))
		json.NewEncoder(response).Encode(`*** Update user profile status ***`)
		span.SetStatus(codes.Ok, "")
	}
}

func (h ProfileHAndler) DeleteUser(response http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "ProfileHAndler.DeleteUser")

	response.Header().Set(contentType, applicationJson)

	vars := mux.Vars(r)
	username := vars["username"]

	if username == "" {
		json.NewEncoder(response).Encode(`*** Username is missing in parameters ***`)
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	user := h.Repo.GetRegularUserByUsername(ctx, username)
	if user == (model.User{}) {
		json.NewEncoder(response).Encode(`*** Username is missing in parameters ***`)
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	} else {
		h.Repo.DeleteUser(ctx, username)
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusOK))
		json.NewEncoder(response).Encode(`*** Successful delete user ***`)
		span.SetStatus(codes.Ok, "")
	}
}

func (h ProfileHAndler) CancelUserRegister(response http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "ProfileHAndler.CancelUserRegister")

	response.Header().Set(contentType, applicationJson)

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		response.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(response).Encode(`*** ID is missing in parameters ***`)
	}

	user := h.Repo.GetUserById(ctx, id)
	if user == (model.Users{}) {
		response.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(response).Encode(`*** User with this ID doesn't exist' ***`)
	} else {
		h.Repo.DeleteUserById(ctx, id)
		json.NewEncoder(response).Encode(`*** Registration successfully canceled ***`)
		span.SetStatus(codes.Ok, "")
	}

}

func (h ProfileHAndler) GetListOfUsersByIds(response http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "ProfileHAndler.GetListOfUsersByIds")

	response.Header().Set(contentType, applicationJson)
	contentType := r.Header.Get(contentType)
	mediatype, _, errContentType := mime.ParseMediaType(contentType)

	if errContentType != nil {
		http.Error(response, errContentType.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errContentType.Error())
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if mediatype != applicationJson {
		err := errors.New(expectJson)
		http.Error(response, err.Error(), http.StatusUnsupportedMediaType)
		span.SetStatus(codes.Error, err.Error())
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusUnsupportedMediaType))
		return
	}

	ids, errDecodeBody := DecodeBodyListOfIDs(r.Body)

	if errDecodeBody != nil {
		http.Error(response, errDecodeBody.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errDecodeBody.Error())
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	users, err := h.Repo.GetListOfUsersByIds(ctx, *ids)
	if err != nil {
		http.Error(response, err.Error(), http.StatusUnsupportedMediaType)
		span.SetStatus(codes.Error, err.Error())
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusUnsupportedMediaType))
		return
	} else {
		if users == nil {
			log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusOK))
			json.NewEncoder(response).Encode(`[]`)
			span.SetStatus(codes.Ok, "")
		} else {
			log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusOK))
			json.NewEncoder(response).Encode(users)
			span.SetStatus(codes.Ok, "")
		}
	}
}

func (h ProfileHAndler) GetUserById(response http.ResponseWriter, r *http.Request) {
	ctx, span := h.Tracer.Start(r.Context(), "ProfileHAndler.GetUserById")

	response.Header().Set(contentType, applicationJson)

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		json.NewEncoder(response).Encode(`ID is missing in parameters`)
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	users := h.Repo.GetUserById(ctx, id)
	if users == (model.Users{}) {
		json.NewEncoder(response).Encode(`Doesn't exist user with this id`)
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	} else {
		log.Println(r.RemoteAddr + " " + r.Method + " " + r.RequestURI + " " + strconv.Itoa(http.StatusOK))
		json.NewEncoder(response).Encode(users)
		span.SetStatus(codes.Ok, "")
	}
}

func ExtractTraceInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
