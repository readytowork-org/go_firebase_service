package go_firebase_service

import (
	"context"

	"firebase.google.com/go"
	"google.golang.org/api/option"
)

type appConfigLogger interface {
	Info(args ...interface{})
	Fatalf(template string, args ...interface{})
}

// AppConfig structure
type AppConfig struct {
	Logger  appConfigLogger
	Options *option.ClientOption
}

// AppService structure
type AppService struct {
	*firebase.App
}

// NewFirebaseApp creates new firebase App instance
func NewFirebaseApp(config AppConfig) AppService {
	app, err := firebase.NewApp(context.Background(), nil, *config.Options)
	if err != nil {
		config.Logger.Fatalf("Firebase NewApp: %v", err)
	}

	config.Logger.Info("✅ Firebase App initialized.")
	return AppService{
		App: app,
	}
}
