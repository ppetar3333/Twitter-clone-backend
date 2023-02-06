package main

import (
	"encoding/json"
	"github.com/ppetar33/twitter-api/tweet-microservice/model"
	"io"
)

func DecodeBody(r io.Reader) (*model.Tweet, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt model.Tweet
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}

	return &rt, nil
}
