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

func Borrow(userUUID uuid.UUID, req models.CreateBorrowRequest) (models.BorrowResponse, error) {
	bookUUID := req.BookID // Already uuid.UUID

	// Check if this user already borrowed this book
	_, err := db.Q.FilterBorrowByUserAndBookID(db.Ctx, gen.FilterBorrowByUserAndBookIDParams{
		BookID: pgtype.UUID{Bytes: bookUUID, Valid: true},
		UserID: pgtype.UUID{Bytes: userUUID, Valid: true},
	})
	if err == nil {
		return models.BorrowResponse{}, errors.New("you have already borrowed this book")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return models.BorrowResponse{}, err
	}
	// Parse due date
	dueDate, err := time.Parse(time.RFC3339, req.DueDate)
	if err != nil {
		return models.BorrowResponse{}, err
	}

	// Get book info
	book, err := db.Q.GetBookByID(db.Ctx, pgtype.UUID{Bytes: bookUUID, Valid: true})
	if err != nil {
		return models.BorrowResponse{}, err
	}

	// Create borrow entry
	borrow, err := db.Q.CreateBorrow(db.Ctx, gen.CreateBorrowParams{
		UserID:     pgtype.UUID{Bytes: userUUID, Valid: true},
		BookID:     pgtype.UUID{Bytes: bookUUID, Valid: true},
		DueDate:    pgtype.Timestamp{Time: dueDate, Valid: true},
		ReturnedAt: pgtype.Timestamp{Valid: false},
	})
	if err != nil {
		return models.BorrowResponse{}, err
	}

	// Decrement available copies
	_, err = db.Q.DecrementAvailableCopiesByID(db.Ctx, pgtype.UUID{Bytes: bookUUID, Valid: true})
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
		ID:         borrow.ID.Bytes,
		UserID:     borrow.UserID.Bytes,
		BookID:     borrow.BookID.Bytes,
		BookTitle:  book.Title,
		DueDate:    borrow.DueDate.Time,
		BorrowedAt: borrow.BorrowedAt.Time,
		ReturnedAt: returnedAt,
	}, nil
}

func Return(borrowUUID uuid.UUID ,req models.ReturnBookRequest) (map[string]string, error) {
	// Parse book_id for incrementing copies
	bookUUID, err := uuid.Parse(req.BookID.String())
	if err != nil {
		return nil, err
	}

	// Update returned_at in borrows table
	err = db.Q.UpdateBorrowReturnedAtByID(db.Ctx, pgtype.UUID{Bytes: borrowUUID, Valid: true})
	if err != nil {
		return nil, err
	}

	// Increment available copies of the book
	_, err = db.Q.IncrementAvailableCopiesByID(db.Ctx, pgtype.UUID{Bytes: bookUUID, Valid: true})
	if err != nil {
		return nil, err
	}

	return map[string]string{"message": "Book returned successfully"}, nil
}


