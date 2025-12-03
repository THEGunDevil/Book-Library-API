package service

import (

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype" // FloatToNumeric converts float64 to pgtype.Float8
)

func UUIDToPGType(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: u,
		Valid: true,
	}
}

// Converts string → pgtype.Text
func StringToPGText(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  s != "",
	}
}
func SafeInt(value interface{}) int {
	if value == nil {
		return 0
	}
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}