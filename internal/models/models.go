package models

import (
	"time"
)

type UrlInfo struct {
	Id          int       `json:"id"`
	Url         string    `json:"url"`
	ShortCode   string    `json:"shortcode"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	AccessCount int       `json:"accessed"`
	OwnerID     int       `json:"owner_id"`
}

type User struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type PostRequestJSON struct {
	Url string`json:"url" binding:"required"`
}

type PutRequestJSON struct {
	Url string `json:"url" binding:"required"`
}

type DeleteRequestJSON struct {
	ShortCode string `json:"short_code" binding:"required"`
}

type PostUserRegistration struct {
	Name string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type PostUserLogin struct {
	Email string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type PostRefreshToken struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}