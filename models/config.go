package models

type Config struct {
	Token     string `json:"token"`
	Username  string `json:"username"`
	FirstTime bool   `json:"first_time"`
	UserType  string `json:"user_type"`
}
