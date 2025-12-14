package models

import (
	"context"
	"time"

	"github.com/joshua-takyi/auction/internal/constants"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Profile struct {
	ID            primitive.ObjectID `bson:"_id" json:"_id"`
	UserID        primitive.ObjectID `bson:"userId" json:"userId"`
	Email         string             `bson:"email" json:"email" validate:"required,email"`
	FirstName     string             `bson:"firstName" json:"firstName" validate:"required"`
	LastName      string             `bson:"lastName" json:"lastName" validate:"required"`
	Role          string             `bson:"role" json:"role" validate:"required"`
	IsVerified    bool               `bson:"isVerified" json:"isVerified"`
	Phone         string             `bson:"phone" json:"phone" validate:"required"`
	RecipientCode string             `bson:"recipientCode" json:"recipientCode" validate:"required"`
	CustomerCode  string             `bson:"customerCode" json:"customerCode" validate:"required"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type ProfileInterface interface {
	CreateProfile(ctx context.Context, email string, profile *Profile) error
	GetUserProfileByID(ctx context.Context, id primitive.ObjectID) (*Profile, error)
}

func (mdb *MongodbRepo) CreateProfile(ctx context.Context, email string, profile *Profile) error {

	if mdb.mongodb == nil {
		return constants.ErrNoClient
	}
	user, err := mdb.FindUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	profile.Email = user.Email
	profile.ID = primitive.NewObjectID()
	profile.UserID = user.ID
	profile.Role = constants.RoleUser
	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()

	_, err = mdb.mongodb.Database(string(constants.DbName)).Collection(string(constants.ProfileCollection)).InsertOne(ctx, profile)
	if err != nil {
		return err
	}

	return nil
}

func (mdb *MongodbRepo) GetUserProfileByID(ctx context.Context, id primitive.ObjectID) (*Profile, error) {
	if mdb.mongodb == nil {
		return nil, constants.ErrNoClient
	}
	profile := &Profile{}
	err := mdb.mongodb.Database(string(constants.DbName)).Collection(string(constants.ProfileCollection)).FindOne(ctx, bson.M{"userId": id}).Decode(profile)
	if err != nil {
		return nil, err
	}
	return profile, nil
}
