package repository

import (
	"context"

	"github.com/be-users-metadata-service/internal/infrastructure/entity"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UserRepository handles users table and meta_data JSONB updates.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetMetaData returns current meta_data for a user (for read-modify-write in app layer).
func (r *UserRepository) GetMetaData(ctx context.Context, userID uuid.UUID) (datatypes.JSON, error) {
	var u entity.User
	if err := r.db.WithContext(ctx).Select("meta_data").Where("id = ?", userID).First(&u).Error; err != nil {
		return nil, err
	}
	return u.MetaData, nil
}

// UpdateMetaData sets users.meta_data for the given user (use inside transaction).
func (r *UserRepository) UpdateMetaData(ctx context.Context, userID uuid.UUID, meta datatypes.JSON) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", userID).Update("meta_data", meta).Error
}

// UpdateMetaDataTx updates meta_data using the given transaction.
func (r *UserRepository) UpdateMetaDataTx(tx *gorm.DB, userID uuid.UUID, meta datatypes.JSON) error {
	return tx.Model(&entity.User{}).Where("id = ?", userID).Update("meta_data", meta).Error
}
