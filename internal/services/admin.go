package services

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"appa_subscriptions/internal/domains"
	dbModels "appa_subscriptions/pkg/db/models"
)

type adminService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAdminService creates a new instance of AdminService
func NewAdminService(db *gorm.DB, logger *zap.Logger) domains.AdminService {
	return &adminService{
		db:     db,
		logger: logger,
	}
}

// CheckEmailExists checks if an email exists in the admin table
func (s *adminService) CheckEmailExists(email string) (bool, error) {
	var count int64
	if err := s.db.Model(&dbModels.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		s.logger.Error("failed to check email existence", zap.String("email", email), zap.Error(err))
		return false, err
	}

	return count > 0, nil
}
