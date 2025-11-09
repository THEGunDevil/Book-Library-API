package models

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type CreateBookRequest struct {
    Title         string                `form:"title" json:"title" binding:"required"`
    Author        string                `form:"author" json:"author" binding:"required"`
    PublishedYear int                   `form:"published_year" json:"published_year" binding:"required"`
    Isbn          string                `form:"isbn" json:"isbn" binding:"required"`
    TotalCopies   int                   `form:"total_copies" json:"total_copies" binding:"required"`
    Genre         string                `form:"genre" json:"genre" binding:"required"`       
    Description   string                `form:"description" json:"description" binding:"required"` 
    Image         *multipart.FileHeader `form:"image" json:"image"` // optional file upload
}


type BookResponse struct {
	ID              uuid.UUID `json:"id"`
	Title           string    `json:"title"`
	Author          string    `json:"author"`
	PublishedYear   int32     `json:"published_year"`
	Isbn            string    `json:"isbn"`
	AvailableCopies int32     `json:"available_copies"`
	TotalCopies     int32     `json:"total_copies"`
	Genre           string    `json:"genre"`       // new
	Description     string    `json:"description"` // new
	ImageURL        string    `json:"image_url"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type UpdateBookRequest struct {
    Title           *string               `form:"title" json:"title"`
    Author          *string               `form:"author" json:"author"`
    PublishedYear   *int32                `form:"published_year" json:"published_year"`
    Isbn            *string               `form:"isbn" json:"isbn"`
    TotalCopies     *int32                `form:"total_copies" json:"total_copies"`
    AvailableCopies *int32                `form:"available_copies" json:"available_copies"`
    Genre           *string               `form:"genre" json:"genre"`
    Description     *string               `form:"description" json:"description"`
    Image           *multipart.FileHeader `form:"image" json:"image"` // optional file
}

type SearchBooksParams struct {
	Genre  *string
	Search *string
}
