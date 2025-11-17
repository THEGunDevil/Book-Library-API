package service

import (
	"encoding/json"

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

// MapToJSONB converts a map[string]interface{} to pgtype.JSONB
func MapToJSONBBytes(m map[string]interface{}) ([]byte, error) {
	if m == nil {
		m = map[string]interface{}{} // default to empty object
	}
	return json.Marshal(m)
}

