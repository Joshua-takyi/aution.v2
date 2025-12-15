package models

import (
	storage_go "github.com/supabase-community/storage-go"
	"github.com/supabase-community/supabase-go"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongodbRepo struct {
	mongodb *mongo.Client
}

func MongodbNewRepo(client *mongo.Client) (*MongodbRepo, error) {
	return &MongodbRepo{
		mongodb: client,
	}, nil
}

type SupabaseRepo struct {
	supabase      *supabase.Client
	storage       *storage_go.Client
	url           string
	anonKey       string
	serviceKey    string
	serviceClient *supabase.Client
}

func NewSupabaseRepo(client *supabase.Client, storage *storage_go.Client, url, anonKey, serviceKey string) *SupabaseRepo {
	var serviceClient *supabase.Client
	if serviceKey != "" {
		serviceClient, _ = supabase.NewClient(url, serviceKey, nil)
	}

	return &SupabaseRepo{
		supabase:      client,
		storage:       storage,
		url:           url,
		anonKey:       anonKey,
		serviceKey:    serviceKey,
		serviceClient: serviceClient,
	}
}
