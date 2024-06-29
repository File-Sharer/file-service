package model

import "time"

type User struct {
	ID           string    `json:"id"`
	Login        string    `json:"login"`
	Role         string    `json:"role"`
	DateAdded    time.Time `json:"dateAdded"`
}

type UserRes struct {
	Data  User   `json:"data"`
	Error string `json:"error"`
	Ok    bool   `json:"ok"`
}
