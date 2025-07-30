package model

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type UserRes struct {
	Data  User   `json:"data"`
	Error string `json:"error"`
	Ok    bool   `json:"ok"`
}
