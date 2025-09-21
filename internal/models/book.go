package models

import (
	"mime/multipart"
	"time"
)

type CreateBookRequest struct {
	Title         string                `form:"title" binding:"required"`
	Author        string                `form:"author" binding:"required"`
	PublishedYear int32                 `form:"published_year"`
	Isbn          string                `form:"isbn"`
	Image           *multipart.FileHeader  `form:"image"` // optional file upload
	TotalCopies   int32                 `form:"total_copies"`
}

type BookResponse struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Author          string    `json:"author"`
	PublishedYear   int32     `json:"published_year"`
	Isbn            string    `json:"isbn"`
	AvailableCopies int32     `json:"available_copies"`
	TotalCopies     int32     `json:"total_copies"`
	ImageURL        string    `json:"image_url"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type UpdateBookRequest struct {
	Title           *string `form:"title"`
	Author          *string `form:"author"`
	PublishedYear   *int32  `form:"published_year"`
	Isbn            *string `form:"isbn"`
	TotalCopies     *int32  `form:"total_copies"`
	AvailableCopies *int32  `form:"available_copies"`
	Image           *multipart.FileHeader  `form:"image"` // optional file upload
}

type ReturnBookRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	BookID string `json:"book_id" binding:"required,uuid"`
}
