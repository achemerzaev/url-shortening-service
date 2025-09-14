package models

import "time"

type UrlInfo struct {
	Id int`json:"id"`
	Url string`json:"url"`
	ShortCode string`json:"shortcode"`
	CreatedAt time.Time`json:"createdAt"`
	UpdatedAt time.Time`json:"updatedAt"`
	AccessCount int`json:"Accessed"`
}