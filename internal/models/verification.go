package models

import (
	"context"
	"errors"
	"time"

	"github.com/joshua-takyi/auction/internal/constants"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Verification struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	Token     string             `bson:"token" json:"token" validate:"required"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	ExpiresAt time.Time          `bson:"expiresAt" json:"expiresAt"`
}

type VerificationInterface interface {
	CreateVerification(ctx context.Context, verification *Verification) (*Verification, error)
	UpdateUserVerificationStatus(ctx context.Context, userID primitive.ObjectID, isVerified bool) error
	DeleteVerificationToken(ctx context.Context, token string) error
	FindVerificationToken(ctx context.Context, token string) (*Verification, error)
	FindVerificationByUserID(ctx context.Context, userID primitive.ObjectID) (*Verification, error)
	DeleteExpiredVerificationTokens(ctx context.Context) error
}

func (mdb *MongodbRepo) CreateVerification(ctx context.Context, verification *Verification) (*Verification, error) {
	if mdb.mongodb == nil {
		return nil, errors.New("mongodb is nil")
	}
	//check if user id already exist in collection table
	filter := bson.M{"userId": verification.UserID}
	user := mdb.mongodb.Database(string(constants.DbName)).Collection(string(constants.VerificationCollection)).FindOne(ctx, filter)

	if user.Err() == nil {
		return nil, constants.ErrUserAlreadyExists
	}

	insertedVerification, err := mdb.mongodb.Database(string(constants.DbName)).Collection(string(constants.VerificationCollection)).InsertOne(ctx, verification)
	if err != nil {
		return nil, err
	}
	verification.ID = insertedVerification.InsertedID.(primitive.ObjectID)
	return verification, nil
}

func (mdb *MongodbRepo) DeleteVerificationToken(ctx context.Context, token string) error {
	if mdb.mongodb == nil {
		return errors.New("mongodb is nil")
	}
	filter := bson.M{"token": token}
	_, err := mdb.mongodb.Database(string(constants.DbName)).Collection(string(constants.VerificationCollection)).DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}
func (mdb *MongodbRepo) FindVerificationToken(ctx context.Context, token string) (*Verification, error) {
	if mdb.mongodb == nil {
		return nil, errors.New("mongodb is nil")
	}
	filter := bson.M{"token": token}
	result := mdb.mongodb.Database(string(constants.DbName)).Collection(string(constants.VerificationCollection)).FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	var verification Verification
	if err := result.Decode(&verification); err != nil {
		return nil, err
	}
	return &verification, nil
}

func (mdb *MongodbRepo) DeleteExpiredVerificationTokens(ctx context.Context) error {
	if mdb.mongodb == nil {
		return errors.New("mongodb is nil")
	}
	filter := bson.M{"expiresAt": bson.M{"$lt": time.Now()}}
	_, err := mdb.mongodb.Database(string(constants.DbName)).Collection(string(constants.VerificationCollection)).DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}

func (mdb *MongodbRepo) FindVerificationByUserID(ctx context.Context, userID primitive.ObjectID) (*Verification, error) {
	if mdb.mongodb == nil {
		return nil, errors.New("mongodb is nil")
	}
	filter := bson.M{"userId": userID}
	result := mdb.mongodb.Database(string(constants.DbName)).Collection(string(constants.VerificationCollection)).FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, result.Err()
	}
	var verification Verification
	if err := result.Decode(&verification); err != nil {
		return nil, err
	}
	return &verification, nil
}
