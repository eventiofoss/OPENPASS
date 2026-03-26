package api

import (
	"github.com/eventiofoss/eventio/backend/internal/service"
	"gorm.io/gorm"
)

// Handler holds dependencies for all request handlers.
type Handler struct {
	DB           *gorm.DB
	Auth         *service.AuthService
	Events       *service.EventService
	Forms        *service.FormService
	Registration *service.RegistrationService
	Analytics    *service.AnalyticsService
	QR           *service.QRService
}
