package dto

import "time"

type HandlerPostRequest struct {
	URL string `json:"url" binding:"required"`
}

type HandlerPostResponse struct {
	URL string `json:"url" binding:"required"`
}

type HandlerGetResponse struct {
	URL string `json:"url" binding:"required"`
}

type HandlerPutRequest struct {
	URL string `json:"url" binding:"required"`
}

type HandlerPutResponse struct {
	Id          int       `json:"id"`
	Url         string    `json:"url"`
	ShortCode   string    `json:"shortcode"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	AccessCount int       `json:"accessed"`
	OwnerID     int       `json:"owner_id"`
}

type HandlerGetStatsResponse struct {
	Id          int       `json:"id"`
	Url         string    `json:"url"`
	ShortCode   string    `json:"shortcode"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	AccessCount int       `json:"accessed"`
	OwnerID     int       `json:"owner_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
