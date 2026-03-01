package ports

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UserRepository reads and updates user metadata (e.g. clas_users.meta_data).
type UserRepository interface {
	GetMetaData(ctx context.Context, userID uuid.UUID) (datatypes.JSON, error)
	UpdateMetaData(ctx context.Context, userID uuid.UUID, meta datatypes.JSON) error
	UpdateMetaDataTx(tx *gorm.DB, userID uuid.UUID, meta datatypes.JSON) error
}
