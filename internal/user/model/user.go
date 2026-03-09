package model

import "time"

// User represents an authenticated customer or admin in the system.
type User struct {
	ID           string    `json:"id"            db:"id"`
	Email        string    `json:"email"         db:"email"`
	PasswordHash string    `json:"-"             db:"password_hash"`
	FirstName    string    `json:"first_name"    db:"first_name"`
	LastName     string    `json:"last_name"     db:"last_name"`
	Role         string    `json:"role"          db:"role"`
	CreatedAt    time.Time `json:"created_at"    db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"    db:"updated_at"`
}

// Address represents a shipping or billing address tied to a user.
type Address struct {
	ID        string    `json:"id"         db:"id"`
	UserID    string    `json:"user_id"    db:"user_id"`
	Label     string    `json:"label"      db:"label"`
	Street    string    `json:"street"     db:"street"`
	City      string    `json:"city"       db:"city"`
	State     string    `json:"state"      db:"state"`
	Zip       string    `json:"zip"        db:"zip"`
	Country   string    `json:"country"    db:"country"`
	IsDefault bool      `json:"is_default" db:"is_default"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// RegisterRequest is the payload for user registration.
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// LoginRequest is the payload for user authentication.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is returned after a successful authentication.
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
}

// UpdateProfileRequest is the payload for updating user profile fields.
type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// CreateAddressRequest is the payload for adding a new address.
type CreateAddressRequest struct {
	Label   string `json:"label"`
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
	Country string `json:"country"`
}
