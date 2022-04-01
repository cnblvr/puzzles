package app

type CookieNotification struct {
	Type    NotificationType `json:"type"`
	Message string           `json:"message"`
}

type NotificationType string

const (
	NotificationSuccess NotificationType = "success"
	NotificationError   NotificationType = "error"
	NotificationWarning NotificationType = "warning"
)
