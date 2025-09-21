package service

import (
	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
)

func AddBook(req models.CreateBookRequest, imageURL string) (models.BookResponse, error) {
	book, err := db.Q.CreateBook(db.Ctx, gen.CreateBookParams{
		Title:  req.Title,
		Author: req.Author,
		PublishedYear: pgtype.Int4{
			Int32: int32(req.PublishedYear),
			Valid: true,
		},
		Isbn: pgtype.Text{
			String: req.Isbn,
			Valid:  true,
		},
		TotalCopies: int32(req.TotalCopies),
		AvailableCopies: pgtype.Int4{
			Int32: int32(req.TotalCopies),
			Valid: true,
		},
		ImageUrl: imageURL,
	})

	if err != nil {
		return models.BookResponse{}, err
	}

	return models.BookResponse{
		ID:              book.ID.String(),
		Title:           book.Title,
		Author:          book.Author,
		PublishedYear:   int32(book.PublishedYear.Int32),
		Isbn:            book.Isbn.String,
		AvailableCopies: int32(book.AvailableCopies.Int32),
		TotalCopies:     int32(book.TotalCopies),
		ImageURL:        book.ImageUrl, // 👈 map back
		CreatedAt:       book.CreatedAt.Time,
		UpdatedAt:       book.UpdatedAt.Time,
	}, nil
}
