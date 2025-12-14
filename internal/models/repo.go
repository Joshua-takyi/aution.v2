package models

import (
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
	supabase *supabase.Client
}

func NewSupabasRepo(client *supabase.Client) (*SupabaseRepo, error) {
	return &SupabaseRepo{
		supabase: client,
	}, nil
}
