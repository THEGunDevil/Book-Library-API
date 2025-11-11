package models

type Profile struct {
	UserName string           `json:"user_name"`
	User     []UserResponse   `json:"user"`
	Reviews  []ReviewResponse `json:"reviews,omitempty"`
	Borrows  []BorrowResponse `json:"borrows,omitempty"` // Borrow history
}
