package model

type User struct {
	ID           string    `json:"id"`
	Role         string    `json:"role"`
}

type UserRes struct {
	Data  User   `json:"data"`
	Error string `json:"error"`
	Ok    bool   `json:"ok"`
}
