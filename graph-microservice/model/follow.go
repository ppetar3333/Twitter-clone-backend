package model

type Follow struct {
	RequestedBy string `json:"requestedBy"`
	RequestedTo string `json:"requestedTo"`
}
