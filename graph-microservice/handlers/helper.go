package handlers

import (
	"encoding/json"
	"github.com/ppetar33/twitter-api/graph-microservice/model"
	"io"
	"net/http"
)

func DecodeEntity(r io.Reader) (*model.Entity, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var entity model.Entity
	if err := dec.Decode(&entity); err != nil {
		return nil, err
	}
	return &entity, nil
}

func DecodeFollow(r io.Reader) (*model.Follow, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var follow model.Follow
	if err := dec.Decode(&follow); err != nil {
		return nil, err
	}
	return &follow, nil
}

func DecodeUnfollow(r io.Reader) (*model.Unfollow, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var unfollow model.Unfollow
	if err := dec.Decode(&unfollow); err != nil {
		return nil, err
	}
	return &unfollow, nil
}

func DecodeRespond(r io.Reader) (*model.Respond, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var respond model.Respond
	if err := dec.Decode(&respond); err != nil {
		return nil, err
	}
	return &respond, nil
}

func DecodeUser(r io.Reader) (*model.User, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var user model.User
	if err := dec.Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func renderJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
