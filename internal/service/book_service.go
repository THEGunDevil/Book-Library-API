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
			Valid: req.PublishedYear != 0,
		},
		Isbn: pgtype.Text{
			String: req.Isbn,
			Valid:  len(req.Isbn) > 0,
		},
		TotalCopies: int32(req.TotalCopies),
		AvailableCopies: pgtype.Int4{
			Int32: int32(req.TotalCopies),
			Valid: true,
		},
		ImageUrl:    imageURL,
		Genre:       req.Genre,
		Description: req.Description,
	})

	if err != nil {
		return models.BookResponse{}, err
	}

	return models.BookResponse{
		ID:              book.ID.Bytes,
		Title:           book.Title,
		Author:          book.Author,
		PublishedYear:   book.PublishedYear.Int32,
		Isbn:            book.Isbn.String,
		AvailableCopies: book.AvailableCopies.Int32,
		TotalCopies:     book.TotalCopies,
		Genre:           book.Genre,
		Description:     book.Description,
		ImageURL:        book.ImageUrl,
		CreatedAt:       book.CreatedAt.Time,
		UpdatedAt:       book.UpdatedAt.Time,
	}, nil
}
