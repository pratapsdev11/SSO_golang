// internal/models/user.go
package models

import (
    "gorm.io/gorm"
    "time"
)

type User struct {
    gorm.Model
    Email     string    `json:"email" gorm:"unique;not null"`
    Name      string    `json:"name"`
    EntraID   string    `json:"entra_id" gorm:"unique"`
    LastLogin time.Time `json:"last_login"`
}

