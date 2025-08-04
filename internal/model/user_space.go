package model

import "time"

type UserSpace struct {
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	Level     uint8     `json:"level"`
	CreatedAt time.Time `json:"createdAt"`
}

type FullUserSpace struct {
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	Level     uint8     `json:"level"`
	CreatedAt time.Time `json:"createdAt"`
	Size      int64     `json:"size"`
}
