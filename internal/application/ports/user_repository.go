package ports

import (
	"context"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UserRepository reads and updates user metadata (clas_users in classified8 DB).
// userID is a string (numeric or UUID) matching the user_id column.
type UserRepository interface {
	GetMetaData(ctx context.Context, userID string) (datatypes.JSON, error)
	UpdateMetaData(ctx context.Context, userID string, meta datatypes.JSON) error
	UpdateMetaDataTx(tx *gorm.DB, userID string, meta datatypes.JSON) error
}
