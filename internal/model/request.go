package model

type PingRequest struct {
	Data string `json:"data" validate:"required"`
}
