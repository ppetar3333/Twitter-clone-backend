package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/ppetar33/twitter-api/model"
	database "github.com/ppetar33/twitter-api/server"
	validation "github.com/ppetar33/twitter-api/validation"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const applicationJson = "application/json"
const expectAppJson = "expect application/json Content-Type"
const contentTypeMessage = "Content-Type"

type AuthHandler struct {
	Tracer trace.Tracer
	Repo   database.AuthRepository
}

func ValidateRequest(response http.ResponseWriter, request *http.Request) bool {
	response.Header().Set(contentTypeMessage, applicationJson)
	contentType := request.Header.Get(contentTypeMessage)
	mediatype, _, errContentType := mime.ParseMediaType(contentType)

	if errContentType != nil {
		http.Error(response, errContentType.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return true
	}

	if mediatype != applicationJson {
		err := errors.New(expectAppJson)
		http.Error(response, err.Error(), http.StatusUnsupportedMediaType)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusUnsupportedMediaType))
		return true
	}
	return false
}

func (h AuthHandler) UserSignupRegular(response http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "AuthHandler.UserSignupRegular")
	defer span.End()

	if ValidateRequest(response, request) {
		return
	}

	user, errDecodeBody := DecodeBodyRegisterRegular(request.Body)

	if errDecodeBody != nil {
		http.Error(response, errDecodeBody.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errDecodeBody.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if user != nil {
		validationErr := validation.SignUpValidationRegularUser(user)

		if validationErr != "" {
			json.NewEncoder(response).Encode(validationErr)
			span.SetStatus(codes.Error, validationErr)
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
			return
		} else {
			authResult, err := h.Repo.SaveCredentialsRegularIntoAuth(ctx, user)

			if authResult == nil {
				json.NewEncoder(response).Encode(err.Error())
				span.SetStatus(codes.Error, err.Error())
				log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
				return
			}

			if err != nil {
				span.SetStatus(codes.Error, err.Error())
				http.Error(response, err.Error(), http.StatusBadRequest)
				log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
				return
			}

			span.SetStatus(codes.Ok, "")
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
			json.NewEncoder(response).Encode(authResult)
		}
	}
}

func ValidateCode(response http.ResponseWriter, request *http.Request) {

	response.Header().Set(contentTypeMessage, applicationJson)
	contentType := request.Header.Get(contentTypeMessage)
	mediatype, _, errContentType := mime.ParseMediaType(contentType)

	if errContentType != nil {
		http.Error(response, errContentType.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if mediatype != applicationJson {
		err := errors.New(expectAppJson)
		http.Error(response, err.Error(), http.StatusUnsupportedMediaType)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusUnsupportedMediaType))
		return
	}

	user, errDecodeBody := DecodeBodyLogin(request.Body)

	if errDecodeBody != nil {
		http.Error(response, errDecodeBody.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if user != nil {
		authResult, err := database.ValidateCode(user)

		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
			return
		}

		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
		json.NewEncoder(response).Encode(authResult)
	}
}

func UserSignupBusiness(response http.ResponseWriter, request *http.Request) {
	if ValidateRequest(response, request) {
		return
	}

	user, errDecodeBody := DecodeBodyRegisterBusiness(request.Body)

	if errDecodeBody != nil {
		http.Error(response, errDecodeBody.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if user != nil {
		validationErr := validation.SignUpValidationBusinessUser(user)
		fmt.Print("VALIDATION ERROR " + validationErr)
		if validationErr != "" {
			json.NewEncoder(response).Encode(validationErr)
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
			return
		} else {
			authResult, err := database.SaveCredentialsBusinessIntoAuth(user)

			if authResult == nil {
				json.NewEncoder(response).Encode(err.Error())
				log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
				return
			}

			if err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
				return
			}

			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
			json.NewEncoder(response).Encode(authResult)
		}
	}
}

func UserLogin(response http.ResponseWriter, request *http.Request) {
	response.Header().Set(contentTypeMessage, applicationJson)
	contentType := request.Header.Get(contentTypeMessage)
	mediatype, _, errContentType := mime.ParseMediaType(contentType)

	if errContentType != nil {
		http.Error(response, errContentType.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if mediatype != applicationJson {
		err := errors.New(expectAppJson)
		http.Error(response, err.Error(), http.StatusUnsupportedMediaType)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusUnsupportedMediaType))
		return
	}

	user, errDecodeBody := DecodeBodyLogin(request.Body)

	if errDecodeBody != nil {
		http.Error(response, errDecodeBody.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if user != nil {
		authResult, err := database.Login(user)

		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
			return
		}

		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
		json.NewEncoder(response).Encode(authResult)
	}
}

func ChangePassword(response http.ResponseWriter, r *http.Request) {

	response.Header().Set(contentTypeMessage, applicationJson)
	contentType := r.Header.Get(contentTypeMessage)
	mediatype, _, errContentType := mime.ParseMediaType(contentType)

	if errContentType != nil {
		http.Error(response, errContentType.Error(), http.StatusBadRequest)
		return
	}

	if mediatype != applicationJson {
		err := errors.New(expectAppJson)
		http.Error(response, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	vars := mux.Vars(r)
	username, ok := vars["username"]
	if !ok {
		json.NewEncoder(response).Encode(`Username is missing in parameters`)
	}

	changePassword, errDecodeBody := DecodeBodyChangePassword(r.Body)

	if errDecodeBody != nil {
		http.Error(response, errDecodeBody.Error(), http.StatusBadRequest)
		return
	}

	if changePassword != nil {
		isValid, errMessage := validation.IsChangePasswordValid(*changePassword)
		if !isValid {
			response.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(response).Encode(errMessage)
		} else {
			err, errString := database.ChangePassword(*changePassword, username)
			if err != nil {
				response.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(response).Encode(err)
			}
			if errString != "" {
				response.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(response).Encode(errString)
			} else {
				json.NewEncoder(response).Encode(`Successful change password!`)
			}
		}
	}

}

func RecoveryPassword(response http.ResponseWriter, r *http.Request) {
	if ValidateRequest(response, r) {
		return
	}

	recoveryPassword, errDecodeBody := DecodeBodyRecoveryPassword(r.Body)

	if errDecodeBody != nil {
		http.Error(response, errDecodeBody.Error(), http.StatusBadRequest)
		return
	}

	if recoveryPassword != nil {
		isValid, errMessage := validation.IsRecoveryPasswordValid(*recoveryPassword)

		if !isValid {
			response.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(response).Encode(errMessage)
		} else {
			err, errString := database.RecoveryPassword(*recoveryPassword)
			if err != nil {
				response.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(response).Encode(err)
			}
			if errString != "" {
				response.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(response).Encode(errString)
			} else {
				json.NewEncoder(response).Encode(`Successfully password recovered!`)
			}
		}
	}

}

func CodeRecovery(response http.ResponseWriter, request *http.Request) {
	if ValidateRequest(response, request) {
		return
	}

	emailBody, errDecodeBody := DecodeBodyCodeRecovery(request.Body)

	if errDecodeBody != nil {
		http.Error(response, errDecodeBody.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if emailBody != nil {
		isValid, errMessage := validation.EmailValidation(*emailBody)

		if !isValid {
			response.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(response).Encode(errMessage)
			return
		}

		userExists := database.GetUserByEmail(*emailBody)

		if userExists == (model.Auth{}) {
			response.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(response).Encode(`User doesn't exists`)
			return
		}

		emailSent, emailSentErr := database.SendCodeViaEmail(*emailBody)

		if emailSentErr != "" {
			response.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(response).Encode(emailSentErr)
			return
		}

		response.WriteHeader(http.StatusOK)
		json.NewEncoder(response).Encode(emailSent)
	}
}

func ExtractTraceInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
