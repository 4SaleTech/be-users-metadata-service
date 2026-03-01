package repository

import (
	"context"

	"github.com/be-users-metadata-service/internal/infrastructure/entity"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UserRepository handles the users/clas_users table and meta_data JSONB updates.
// Uses the configured table name (e.g. clas_users) on the users DB connection.
type UserRepository struct {
	db        *gorm.DB
	tableName string
}

// NewUserRepository creates a new UserRepository. tableName is the table to use (e.g. "clas_users"); if empty, "users" is used.
func NewUserRepository(db *gorm.DB, tableName string) *UserRepository {
	if tableName == "" {
		tableName = "users"
	}
	return &UserRepository{db: db, tableName: tableName}
}

func (r *UserRepository) table(db *gorm.DB) *gorm.DB {
	return db.Table(r.tableName).Model(&entity.User{})
}

// GetMetaData returns current meta_data for a user (for read-modify-write in app layer).
func (r *UserRepository) GetMetaData(ctx context.Context, userID uuid.UUID) (datatypes.JSON, error) {
	var u entity.User
	if err := r.table(r.db.WithContext(ctx)).Select("meta_data").Where("id = ?", userID).First(&u).Error; err != nil {
		return nil, err
	}
	return u.MetaData, nil
}

// UpdateMetaData sets meta_data for the given user on the users DB (separate from primary DB transaction).
func (r *UserRepository) UpdateMetaData(ctx context.Context, userID uuid.UUID, meta datatypes.JSON) error {
	return r.table(r.db.WithContext(ctx)).Where("id = ?", userID).Update("meta_data", meta).Error
}

// UpdateMetaDataTx updates meta_data using the given transaction (only valid when tx is the users DB; used in single-DB mode).
func (r *UserRepository) UpdateMetaDataTx(tx *gorm.DB, userID uuid.UUID, meta datatypes.JSON) error {
	return r.table(tx).Where("id = ?", userID).Update("meta_data", meta).Error
}
