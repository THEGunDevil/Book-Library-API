package models

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type CreateBookRequest struct {
	Title         string                `form:"title" json:"title"`
	Author        string                `form:"author" json:"author"`
	PublishedYear int                   `form:"published_year" json:"published_year"`
	Isbn          string                `form:"isbn" json:"isbn"`
	TotalCopies   int                   `form:"total_copies" json:"total_copies"`
	Genre         string                `form:"genre" json:"genre"`             // new
	Description   string                `form:"description" json:"description"` // new
	Image         *multipart.FileHeader `form:"image"`
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
	Title           *string               `form:"title"`
	Author          *string               `form:"author"`
	PublishedYear   *int32                `form:"published_year"`
	Isbn            *string               `form:"isbn"`
	TotalCopies     *int32                `form:"total_copies"`
	AvailableCopies *int32                `form:"available_copies"`
	Genre           *string               `form:"genre"`       // new
	Description     *string               `form:"description"` // new
	Image           *multipart.FileHeader `form:"image"`       // optional file upload
}
type SearchBooksParams struct {
    Genre  *string
    Search *string
}



