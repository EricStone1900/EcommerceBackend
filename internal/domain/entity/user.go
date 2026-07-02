package entity

import "time"

// Role represents a user role for RBAC.
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleMember   Role = "member"
	RoleCustomer Role = "customer"
)

// User is the domain entity for an authenticated user.
// Soft delete (deleted_at) is handled at the infrastructure/GORM layer.
type User struct {
	ID           uint      `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
