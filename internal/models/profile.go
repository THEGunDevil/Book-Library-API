package models

type Profile struct {
	FullName string           `json:"full_name"`
	User     []UserResponse   `json:"user_info"`
	Reviews  []ReviewResponse `json:"reviews,omitempty"`
	Borrows  []BorrowResponse `json:"borrows,omitempty"` // Borrow history
}
