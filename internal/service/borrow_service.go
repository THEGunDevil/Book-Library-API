package service

import (
	"errors"
	"time"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func Borrow(req models.CreateBorrowRequest) (models.BorrowResponse, error) {
	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return models.BorrowResponse{}, err
	}
	bookUUID, err := uuid.Parse(req.BookID)
	if err != nil {
		return models.BorrowResponse{}, err
	}
	_, err = db.Q.FilterBorrowByUserAndBookID(db.Ctx, gen.FilterBorrowByUserAndBookIDParams{
		UserID: pgtype.UUID{Bytes: userUUID, Valid: true},
		BookID: pgtype.UUID{Bytes: bookUUID, Valid: true},
	})
	if err == nil {
		return models.BorrowResponse{}, errors.New("already borrowed this book")
	}

	dueDate, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		return models.BorrowResponse{}, err
	}
	borrow, err := db.Q.CreateBorrow(db.Ctx, gen.CreateBorrowParams{
		UserID:     pgtype.UUID{Bytes: userUUID, Valid: true},
		BookID:     pgtype.UUID{Bytes: bookUUID, Valid: true},
		DueDate:    pgtype.Timestamp{Time: dueDate, Valid: true},
		ReturnedAt: pgtype.Timestamp{Valid: false},
	})
	if err != nil {
		return models.BorrowResponse{}, err
	}
	_, err = db.Q.DecrementAvailableCopiesByID(db.Ctx, pgtype.UUID{
		Bytes: bookUUID,
		Valid: true,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.BorrowResponse{}, errors.New("no copies available")
		}
		return models.BorrowResponse{}, err
	}
	var returnedAt *time.Time
	if borrow.ReturnedAt.Valid {
		returnedAt = &borrow.ReturnedAt.Time
	}
	return models.BorrowResponse{
		ID:         borrow.ID.String(),
		UserID:     borrow.UserID.String(),
		BookID:     borrow.BookID.String(),
		DueDate:    borrow.DueDate.Time,
		BorrowedAt: borrow.BorrowedAt.Time,
		ReturnedAt: returnedAt,
	}, nil
}

func Return(req models.ReturnBookRequest) (map[string]string, error) {

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, err
	}
	bookUUID, err := uuid.Parse(req.BookID)
	if err != nil {
		return nil, err
	}
	err = db.Q.UpdateBorrowByUserAndBookID(db.Ctx, gen.UpdateBorrowByUserAndBookIDParams{
		UserID: pgtype.UUID{Bytes: userUUID, Valid: true},
		BookID: pgtype.UUID{Bytes: bookUUID, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	_, err = db.Q.IncrementAvailableCopiesByID(db.Ctx, pgtype.UUID{Bytes: bookUUID, Valid: true})
	if err != nil {
		return nil, err
	}
	return map[string]string{"message": "Book returned successfully"}, nil
}