package go_firebase_service

import (
	"context"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go"
)

type storeClientLogger interface {
	Fatalf(template string, args ...interface{})
}

type StoreClientConfig struct {
	Logger storeClientLogger
	App    *firebase.App
}

type StoreClientService struct {
	*firestore.Client
}

// NewFireStoreClient creates new firestore client
func NewFireStoreClient(config StoreClientConfig) StoreClientService {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	firestoreClient, err := config.App.Firestore(ctx)
	if err != nil {
		config.Logger.Fatalf("Firestore client: %v", err)
	}

	return StoreClientService{Client: firestoreClient}
}
