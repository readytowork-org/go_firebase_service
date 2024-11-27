package go_firebase_service

import (
	"context"

	"firebase.google.com/go"
	"firebase.google.com/go/messaging"
)

type loggerCMClient interface {
	Fatalf(template string, args ...interface{})
}

type CMClientConfig struct {
	Logger loggerCMClient
	App    *firebase.App
}

type CMClientService struct {
	*messaging.Client
}

// NewFirebaseCMClient creates new firebase cloud messaging client
func NewFirebaseCMClient(config CMClientConfig) CMClientService {
	ctx := context.Background()
	messagingClient, err := config.App.Messaging(ctx)
	if err != nil {
		config.Logger.Fatalf("Firebase messaing: %v", err)
	}
	return CMClientService{
		Client: messagingClient,
	}
}
