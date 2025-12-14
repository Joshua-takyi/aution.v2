package models

import (
	"context"
	"errors"
	"time"

	"github.com/joshua-takyi/auction/internal/constants"
	"github.com/joshua-takyi/auction/internal/helpers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID         primitive.ObjectID `bson:"_id" json:"_id"`
	Email      string             `bson:"email" json:"email" validate:"required,email"`
	Password   string             `bson:"password" json:"password" validate:"required"`
	IsVerified bool               `bson:"isVerified" json:"isVerified"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type RefreshToken struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	Token     string             `bson:"token" json:"token" validate:"required"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	ExpiresAt time.Time          `bson:"expiresAt" json:"expiresAt"`
}

type UserInterface interface {
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, email, password string) (*User, error)
	UpdateUserVerificationStatus(ctx context.Context, userID primitive.ObjectID, isVerified bool) error
	AuthenticateUser(ctx context.Context, email, password string) (*User, error)
	IsVerified(ctx context.Context, email string) (bool, error)
}

func (mdb *MongodbRepo) CreateUser(ctx context.Context, email, password string) (*User, error) {
	if mdb.mongodb == nil {
		return nil, errors.New("mongodb is nil")
	}

	filter := bson.M{"email": email}
	user := mdb.mongodb.Database(string(constants.DbName)).Collection(string(constants.UserCollection)).FindOne(ctx, filter)
	if user.Err() == nil {
		return nil, constants.ErrUserAlreadyExists
	}

	hashedPassword, err := helpers.HashPassword(password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	newUser := User{
		ID:         primitive.NewObjectID(),
		Email:      email,
		IsVerified: false,
		Password:   hashedPassword,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	insertedUser, err := mdb.mongodb.Database(string(constants.DbName)).Collection(string(constants.UserCollection)).InsertOne(ctx, newUser)
	if err != nil {
		return nil, err
	}

	newUser.ID = insertedUser.InsertedID.(primitive.ObjectID)

	return &newUser, nil
}

func (mdb *MongodbRepo) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	if mdb.mongodb == nil {
		return nil, errors.New("mongodb is nil")
	}

	filter := bson.M{"email": email}
	user := mdb.mongodb.Database(string(constants.DbName)).Collection(string(constants.UserCollection)).FindOne(ctx, filter)
	if user.Err() != nil {
		return nil, user.Err()
	}

	var result User
	if err := user.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (mdb *MongodbRepo) UpdateUserVerificationStatus(ctx context.Context, userID primitive.ObjectID, verified bool) error {
	if mdb.mongodb == nil {
		return errors.New("mongodb is nil")
	}

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": bson.M{"isVerified": verified}}
	_, err := mdb.mongodb.Database(string(constants.DbName)).Collection(string(constants.UserCollection)).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (mdb *MongodbRepo) AuthenticateUser(ctx context.Context, email, password string) (*User, error) {
	if mdb.mongodb == nil {
		return nil, errors.New("mongodb is nil")
	}

	filter := bson.M{"email": email}
	user := mdb.mongodb.Database(string(constants.DbName)).Collection(string(constants.UserCollection)).FindOne(ctx, filter)
	if user.Err() != nil {
		return nil, user.Err()
	}

	var result User
	if err := user.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (mdb *MongodbRepo) IsVerified(ctx context.Context, email string) (bool, error) {
	if mdb.mongodb == nil {
		return false, errors.New("mongodb is nil")
	}

	filter := bson.M{"email": email}
	user := mdb.mongodb.Database(string(constants.DbName)).Collection(string(constants.UserCollection)).FindOne(ctx, filter)
	if user.Err() != nil {
		return false, user.Err()
	}

	var result User
	if err := user.Decode(&result); err != nil {
		return false, err
	}

	return result.IsVerified, nil
}
