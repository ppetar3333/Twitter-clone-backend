package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"mime"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/ppetar33/twitter-api/graph-microservice/data"
	"github.com/ppetar33/twitter-api/graph-microservice/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type GraphHandler struct {
	Tracer trace.Tracer
	Repo   data.GraphRepository
}

const contentTypeText = "Content-Type"
const applicationJsonText = "application/json"
const errorMessageText = "expect application/json Content-Type"
const notValidRequest = "Request is not valid"

func (h GraphHandler) RegisterUser(writer http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "GraphHandler.RegisterUser")

	writer.Header().Set(contentTypeText, applicationJsonText)

	contentType := request.Header.Get(contentTypeText)
	mediatype, _, errContentType := mime.ParseMediaType(contentType)

	if errContentType != nil {
		http.Error(writer, errContentType.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errContentType.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if mediatype != applicationJsonText {
		err := errors.New(errorMessageText)
		http.Error(writer, err.Error(), http.StatusUnsupportedMediaType)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusUnsupportedMediaType))
		return
	}

	entity, errDecodeBody := DecodeEntity(request.Body)

	if errDecodeBody != nil {
		http.Error(writer, errContentType.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errDecodeBody.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if entity.Username == "" ||
		entity.Name == "" ||
		entity.ID == "" ||
		(entity.Type != "regular" && entity.Type != "business") {
		http.Error(writer, "User is not valid", http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	} else {
		result, err := h.Repo.AddEntityInDatabase(ctx, entity)

		if err != nil {
			http.Error(writer, errContentType.Error(), http.StatusBadRequest)
			span.SetStatus(codes.Error, err.Error())
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
			return
		}
		renderJSON(writer, result)
		span.SetStatus(codes.Ok, "")
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
	}
}

func (h GraphHandler) Follow(writer http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "GraphHandler.Follow")

	writer.Header().Set(contentTypeText, applicationJsonText)

	contentType := request.Header.Get(contentTypeText)
	mediatype, _, errContentType := mime.ParseMediaType(contentType)

	if errContentType != nil {
		http.Error(writer, errContentType.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errContentType.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if mediatype != applicationJsonText {
		err := errors.New(errorMessageText)
		http.Error(writer, err.Error(), http.StatusUnsupportedMediaType)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusUnsupportedMediaType))
		return
	}

	follow, errDecodeBody := DecodeFollow(request.Body)

	if errDecodeBody != nil {
		http.Error(writer, errContentType.Error(), http.StatusBadRequest)
		span.SetStatus(codes.Error, errDecodeBody.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if follow.RequestedTo == "" || follow.RequestedBy == "" {
		http.Error(writer, notValidRequest, http.StatusBadRequest)
	} else {
		result, err := h.Repo.AddFollowRelationInDatabase(ctx, follow)
		if err != nil {
			http.Error(writer, errContentType.Error(), http.StatusBadRequest)
			span.SetStatus(codes.Error, err.Error())
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
			return
		}
		renderJSON(writer, result)
		span.SetStatus(codes.Ok, "")
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
	}
}

func (h GraphHandler) RequestFollow(writer http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "GraphHandler.RequestFollow")

	writer.Header().Set(contentTypeText, applicationJsonText)

	contentType := request.Header.Get(contentTypeText)
	mediatype, _, errContentType := mime.ParseMediaType(contentType)

	if errContentType != nil {
		http.Error(writer, errContentType.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if mediatype != applicationJsonText {
		err := errors.New(errorMessageText)
		http.Error(writer, err.Error(), http.StatusUnsupportedMediaType)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusUnsupportedMediaType))
		return
	}

	follow, errDecodeBody := DecodeFollow(request.Body)

	if errDecodeBody != nil {
		http.Error(writer, errContentType.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if follow.RequestedTo == "" || follow.RequestedBy == "" {
		http.Error(writer, errContentType.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	} else {
		result, err := h.Repo.AddRequestFollowRelationInDatabase(ctx, follow)
		if err != nil {
			http.Error(writer, errContentType.Error(), http.StatusBadRequest)
			span.SetStatus(codes.Error, err.Error())
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
			return
		}
		renderJSON(writer, result)
		span.SetStatus(codes.Ok, "")
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
	}
}

func (h GraphHandler) Unfollow(writer http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "GraphHandler.Unfollow")

	writer.Header().Set(contentTypeText, applicationJsonText)

	contentType := request.Header.Get(contentTypeText)
	mediatype, _, errContentType := mime.ParseMediaType(contentType)

	if errContentType != nil {
		http.Error(writer, errContentType.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if mediatype != applicationJsonText {
		err := errors.New(errorMessageText)
		http.Error(writer, err.Error(), http.StatusUnsupportedMediaType)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusUnsupportedMediaType))
		return
	}

	unfollow, errDecodeBody := DecodeUnfollow(request.Body)

	if errDecodeBody != nil {
		http.Error(writer, errDecodeBody.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if unfollow.WantToUnfollow == "" || unfollow.User == "" {
		http.Error(writer, notValidRequest, http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
	} else {
		result, err := h.Repo.RemoveFollowRelationInDatabase(ctx, unfollow)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			span.SetStatus(codes.Error, err.Error())
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
			return
		}
		renderJSON(writer, result)
		span.SetStatus(codes.Ok, "")
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
	}
}

func (h GraphHandler) DeclineRequest(writer http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "GraphHandler.DeclineRequest")

	writer.Header().Set(contentTypeText, applicationJsonText)

	contentType := request.Header.Get(contentTypeText)
	mediatype, _, errContentType := mime.ParseMediaType(contentType)

	if errContentType != nil {
		http.Error(writer, errContentType.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if mediatype != applicationJsonText {
		err := errors.New(errorMessageText)
		http.Error(writer, err.Error(), http.StatusUnsupportedMediaType)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusUnsupportedMediaType))
		return
	}

	respond, errDecodeBody := DecodeRespond(request.Body)

	if errDecodeBody != nil {
		http.Error(writer, errDecodeBody.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if respond.OnRequestFromUser == "" || respond.User == "" {
		http.Error(writer, notValidRequest, http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
	} else {
		result, err := h.Repo.DeleteFollowRequestRelationInDatabase(ctx, respond)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			span.SetStatus(codes.Error, err.Error())
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
			return
		}
		renderJSON(writer, result)
		span.SetStatus(codes.Ok, "")
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
	}
}

func (h GraphHandler) AcceptRequest(writer http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "GraphHandler.AcceptRequest")

	writer.Header().Set(contentTypeText, applicationJsonText)

	contentType := request.Header.Get(contentTypeText)
	mediatype, _, errContentType := mime.ParseMediaType(contentType)

	if errContentType != nil {
		http.Error(writer, errContentType.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusUnsupportedMediaType))
		return
	}

	if mediatype != applicationJsonText {
		err := errors.New(errorMessageText)
		http.Error(writer, err.Error(), http.StatusUnsupportedMediaType)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusUnsupportedMediaType))
		return
	}

	respond, errDecodeBody := DecodeRespond(request.Body)

	if errDecodeBody != nil {
		http.Error(writer, errDecodeBody.Error(), http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
		return
	}

	if respond.OnRequestFromUser == "" || respond.User == "" {
		http.Error(writer, notValidRequest, http.StatusBadRequest)
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusBadRequest))
	} else {
		result, err := h.Repo.AcceptFollowRequestRelationInDatabase(ctx, respond)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			span.SetStatus(codes.Error, err.Error())
			log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusInternalServerError))
			return
		}
		renderJSON(writer, result)
		span.SetStatus(codes.Ok, "")
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
	}
}

func (h GraphHandler) GetAllFollowing(writer http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "GraphHandler.GetAllFollowing")

	writer.Header().Set(contentTypeText, applicationJsonText)

	vars := mux.Vars(request)
	id, ok := vars["id"]
	if !ok {
		json.NewEncoder(writer).Encode(`ID is missing in parameters`)
	}
	var user model.User
	user.ID = id

	result, err := h.Repo.GetAllFollowingFromDatabase(ctx, &user)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusInternalServerError))
		return
	} else {
		renderJSON(writer, result)
		span.SetStatus(codes.Ok, "")
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
	}
}

func (h GraphHandler) GetAllFollowers(writer http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "GraphHandler.GetAllFollowers")

	writer.Header().Set(contentTypeText, applicationJsonText)

	vars := mux.Vars(request)
	id, ok := vars["id"]
	if !ok {
		json.NewEncoder(writer).Encode(`ID is missing in parameters`)
	}
	var user model.User
	user.ID = id

	result, err := h.Repo.GetAllFollowersFromDatabase(ctx, &user)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusInternalServerError))
		return
	} else {
		renderJSON(writer, result)
		span.SetStatus(codes.Ok, "")
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
	}

	log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI)
}

func (h GraphHandler) GetAllRecommended(writer http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "GraphHandler.GetAllRecommended")

	writer.Header().Set(contentTypeText, applicationJsonText)

	vars := mux.Vars(request)
	id, ok := vars["id"]
	if !ok {
		json.NewEncoder(writer).Encode(`ID is missing in parameters`)
	}
	var user model.User
	user.ID = id

	result, err := h.Repo.GetAllRecommendedFromDatabase(ctx, &user)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusInternalServerError))
		return
	} else {
		renderJSON(writer, result)
		span.SetStatus(codes.Ok, "")
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
	}
}

func (h GraphHandler) GetAllRequests(writer http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "GraphHandler.GetAllRequests")

	writer.Header().Set(contentTypeText, applicationJsonText)

	vars := mux.Vars(request)
	id, ok := vars["id"]
	if !ok {
		json.NewEncoder(writer).Encode(`ID is missing in parameters`)
	}
	var user model.User
	user.ID = id

	result, err := h.Repo.GetAllRequestsFromDatabase(ctx, &user)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusInternalServerError))
		return
	} else {
		renderJSON(writer, result)
		span.SetStatus(codes.Ok, "")
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
	}
}

func (h GraphHandler) DeleteUser(writer http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "GraphHandler.RegisterUser")

	writer.Header().Set(contentTypeText, applicationJsonText)

	vars := mux.Vars(request)
	id, ok := vars["id"]
	if !ok {
		json.NewEncoder(writer).Encode(`ID is missing in parameters`)
	}
	var user model.User
	user.ID = id
	result, err := h.Repo.DeleteUserFromDatabase(ctx, &user)

	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusInternalServerError))
		return
	} else {
		renderJSON(writer, result)
		span.SetStatus(codes.Ok, "")
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
	}
}

func (h GraphHandler) CheckRelathionship(writer http.ResponseWriter, request *http.Request) {
	ctx, span := h.Tracer.Start(request.Context(), "GraphHandler.CheckRelathionship")

	writer.Header().Set(contentTypeText, applicationJsonText)

	vars := mux.Vars(request)
	requestedBy, ok := vars["requestedBy"]
	if !ok {
		json.NewEncoder(writer).Encode(`requestedBy is missing in parameters`)
	}

	requestedTo, ok := vars["requestedTo"]
	if !ok {
		json.NewEncoder(writer).Encode(`requestedTo is missing in parameters`)
	}

	var follow model.Follow
	follow.RequestedBy = requestedBy
	follow.RequestedTo = requestedTo

	result, err := h.Repo.CheckRelatinosInDatabase(ctx, &follow)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		span.SetStatus(codes.Error, err.Error())
		log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusInternalServerError))
		return
	}
	renderJSON(writer, result)
	span.SetStatus(codes.Ok, "")
	log.Println(request.RemoteAddr + " " + request.Method + " " + request.RequestURI + " " + strconv.Itoa(http.StatusOK))
}

func ExtractTraceInfoMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
