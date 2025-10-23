package models

type Profile struct {
	FullName string           `json:"full_name"`
	User     []UserResponse   `json:"user"`
	Reviews  []ReviewResponse `json:"reviews,omitempty"`
	Borrows  []BorrowResponse `json:"borrows,omitempty"` // Borrow history
}
