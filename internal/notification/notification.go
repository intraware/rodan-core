package notification

import (
	"errors"

	"github.com/intraware/rodan/internal/utils/values"
)

func SendNotification(message string) error {
	cfg := values.GetConfig().App.Notification
	if !cfg.Enabled {
		return errors.New("Notification is not enabled")
	}
	if cfg.DeliveryMethod == "http" {
		return httpNotif(message, cfg)
	} else if cfg.DeliveryMethod == "kafka" {
		return errors.New("Kafka is not implemented")
	} else {
		return errors.New("Invalid delivery method")
	}
}
