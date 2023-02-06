package main

import (
	"encoding/json"
	"io"

	"github.com/ppetar33/twitter-api/model"
)

func DecodeBodyLogin(r io.Reader) (*model.Auth, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt model.Auth
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}
	return &rt, nil
}

func DecodeBodyRegisterRegular(r io.Reader) (*model.User, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt model.User
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}

	return &rt, nil
}

func DecodeBodyRegisterBusiness(r io.Reader) (*model.UserBusiness, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt model.UserBusiness
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}

	return &rt, nil
}

func DecodeBodyChangePassword(r io.Reader) (*model.ChangePassword, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt model.ChangePassword
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}

	return &rt, nil
}

func DecodeBodyRecoveryPassword(r io.Reader) (*model.RecoveryPassword, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt model.RecoveryPassword
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}

	return &rt, nil
}

func DecodeBodyCodeRecovery(r io.Reader) (*model.CodeRecovery, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()

	var rt model.CodeRecovery
	if err := dec.Decode(&rt); err != nil {
		return nil, err
	}

	return &rt, nil
}
