package models

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type CreateBookRequest struct {
    Title         string                `json:"title" form:"title" binding:"required"`           // required
    Author        string                `json:"author" form:"author" binding:"required"`         // required
    PublishedYear int                   `json:"published_year" form:"published_year" binding:"required"` // required
    Isbn          string                `json:"isbn" form:"isbn" binding:"required"`             // required
    TotalCopies   int                   `json:"total_copies" form:"total_copies" binding:"required"` // required
    Genre         string                `json:"genre" form:"genre" binding:"required"`           // required
    Description   string                `json:"description" form:"description" binding:"required"` // required
    Image         *multipart.FileHeader `json:"image" form:"image"`                               // optional file
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
    Title           *string               `json:"title" form:"title"`                       // optional
    Author          *string               `json:"author" form:"author"`                     // optional
    PublishedYear   *int32                `json:"published_year" form:"published_year"`     // optional
    Isbn            *string               `json:"isbn" form:"isbn"`                         // optional
    TotalCopies     *int32                `json:"total_copies" form:"total_copies"`         // optional
    AvailableCopies *int32                `json:"available_copies" form:"available_copies"` // optional
    Genre           *string               `json:"genre" form:"genre"`                       // optional
    Description     *string               `json:"description" form:"description"`           // optional
    Image           *multipart.FileHeader `json:"image" form:"image"`                       // optional file
}


type SearchBooksParams struct {
	Genre  *string
	Search *string
}
