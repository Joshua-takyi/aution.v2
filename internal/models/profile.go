package models

// import (
// 	"context"
// 	"errors"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/joshua-takyi/auction/internal/constants"
// 	"go.mongodb.org/mongo-driver/bson"
// )

// type Profile struct {
// 	ID            uuid.UUID `db:"id" json:"id"`
// 	UserID        uuid.UUID `db:"user_id" json:"user_id"`
// 	Email         string    `db:"email" json:"email" validate:"required,email"`
// 	FirstName     string    `db:"first_name" json:"first_name" validate:"required"`
// 	LastName      string    `db:"last_name" json:"last_name" validate:"required"`
// 	Role          string    `db:"role" json:"role" validate:"required"`
// 	IsVerified    bool      `db:"is_verified" json:"is_verified"`
// 	Phone         string    `db:"phone" json:"phone" validate:"required"`
// 	RecipientCode string    `db:"recipient_code" json:"recipient_code" validate:"required"`
// 	CustomerCode  string    `db:"customer_code" json:"customer_code" validate:"required"`
// 	CreatedAt     time.Time `db:"created_at" json:"created_at"`
// 	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
// }

// type ProfileInterface interface {
// 	CreateProfile(ctx context.Context, email string, profile *Profile) error
// 	GetUserProfileByID(ctx context.Context, id uuid.UUID) (*Profile, error)
// 	GetProfileByEmail(ctx context.Context, email string) (*Profile, error)
// 	UpdateProfile(ctx context.Context, profile *Profile) error
// 	DeleteUserProfile(ctx context.Context, id uuid.UUID) error
// }

// func (mdb *MongodbRepo) CreateProfile(ctx context.Context, email string, profile *Profile) error {
// 	if mdb.mongodb == nil {
// 		return errors.New("mongodb is nil")
// 	}

// 	filter := bson.M{"email": email}
// 	result := mdb.mongodb.Database(string(constants.DbName)).Collection(string(constants.UserCollection)).FindOne(ctx, filter)
// 	if result.Err() != nil {
// 		return result.Err()
// 	}

// 	// For Supabase, we don't necessarily check specific local users, but profile creation might imply it.
// 	// We'll trust the ID passed in profile.ID or generate one if nil, but generally Supabase ID comes from Auth.
// 	if profile.ID == uuid.Nil {
// 		profile.ID = uuid.New()
// 	}

// 	_, err := mdb.mongodb.Database(string(constants.DbName)).Collection(string(constants.ProfileCollection)).InsertOne(ctx, profile)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (mdb *MongodbRepo) GetUserProfileByID(ctx context.Context, id uuid.UUID) (*Profile, error) {
// 	if mdb.mongodb == nil {
// 		return nil, errors.New("mongodb is nil")
// 	}

// 	// Searching by UUID in Mongo? We store it as binary or string depending on driver settings.
// 	// Usually easiest to query as UUID if configured, or we might need to handle BSON wrapping.
// 	// For simplicity, let's assume UUID codec is working or we pass as is.
// 	filter := bson.M{"userId": id}
// 	profile := mdb.mongodb.Database(string(constants.DbName)).Collection(string(constants.ProfileCollection)).FindOne(ctx, filter)
// 	if profile.Err() != nil {
// 		return nil, profile.Err()
// 	}

// 	var result Profile
// 	if err := profile.Decode(&result); err != nil {
// 		return nil, err
// 	}

// 	return &result, nil
// }

// func (mdb *MongodbRepo) GetProfileByEmail(ctx context.Context, email string) (*Profile, error) {
// 	// Stub implementation
// 	return nil, errors.New("not implemented")
// }

// func (mdb *MongodbRepo) UpdateProfile(ctx context.Context, profile *Profile) error {
// 	// Stub implementation
// 	return errors.New("not implemented")
// }

// func (mdb *MongodbRepo) DeleteUserProfile(ctx context.Context, id uuid.UUID) error {
// 	// Stub implementation
// 	return errors.New("not implemented")
// }
