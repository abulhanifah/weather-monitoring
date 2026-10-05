package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User untuk kebutuhan auth JWT
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Name      string    `gorm:"type:varchar(50);not null" json:"name"`
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`
	Role      string    `gorm:"type:varchar(50);not null" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Claims payload JWT, konsisten dengan middleware (claims["email"])
type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// LoginRequest payload login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse response login berisi token
type LoginResponse struct {
	Token string `json:"token"`
}
