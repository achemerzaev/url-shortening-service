package models

import "time"

type UrlInfo struct {
	Id int`json:"id"`
	Url string`json:"url"`
	ShortCode string`json:"shortcode"`
	CreatedAt time.Time`json:"createdAt"`
	UpdatedAt time.Time`json:"updatedAt"`
	AccessCount int`json:"accessed"`
	OwnerID int`json:"owner_id"`
}

type User struct {
	Id int`json:"id"`
	Name string`json:"name"`
	Email string`json:"email"`
	Password string`json:"password"`
}

type Tokens struct {
	AccessToken string`json:"access_token"`
	RefreshToken string`json:"refresh_token"`
}