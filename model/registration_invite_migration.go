package model

import (
	"fmt"

	"gorm.io/gorm"
)

func migrateRegistrationInvites(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("migrate registration invites: database is nil")
	}
	if err := db.AutoMigrate(&RegistrationInvite{}); err != nil {
		return fmt.Errorf("migrate registration invites: %w", err)
	}
	return nil
}
