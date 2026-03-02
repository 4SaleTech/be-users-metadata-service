package entity

import (
	"gorm.io/datatypes"
)

// User represents clas_users (classified8). Only meta_data is read/updated; no updated_at (column may not exist).
type User struct {
	UserID   string         `gorm:"column:user_id;primaryKey"`
	MetaData datatypes.JSON `gorm:"column:meta_data;type:json"`
}

// TableName overrides table name (clas_users in classified8).
func (User) TableName() string { return "clas_users" }
