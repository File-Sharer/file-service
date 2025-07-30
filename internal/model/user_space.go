package model

import "time"

type UserSpace struct {
	UserID    string    `json:"userId"`
	Level     uint8     `json:"level"`
	CreatedAt time.Time `json:"createdAt"`
}
