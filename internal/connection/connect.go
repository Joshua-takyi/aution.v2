package connection

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/resend/resend-go/v3"
	"github.com/supabase-community/supabase-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client         *mongo.Client
	ResendClient   *resend.Client
	SupabaseClient *supabase.Client
)

func Connect(url, pass string) (*mongo.Client, error) {
	if url == "" || pass == "" {
		return nil, constants.ErrEmptyFields
	}
	fullUrl := strings.Replace(url, "<MONGODB_PASSWORD>", pass, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(fullUrl)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func Disconnect() error {
	if Client == nil {
		return constants.ErrNoClient
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return Client.Disconnect(ctx)
}

func ResendConnect(key string) (*resend.Client, error) {
	if key == "" {
		return nil, constants.ErrEmptyFields
	}
	ResendClient = resend.NewClient(key)
	return ResendClient, nil
}

func ConnectSupabase(url, key string) (*supabase.Client, error) {
	client, err := supabase.NewClient(url, key, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to supabase server")
	}
	SupabaseClient = client
	return SupabaseClient, nil
}
