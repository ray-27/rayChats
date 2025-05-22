// models/user.go
package models

import "time"

// import (
// 	"time"
// )

type UserCred struct {
	UUID        string    `json:"uuid"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	PhoneNumber string    `json:"phoneno"`
	CreatedAt   time.Time `json:"created_at"`
}
