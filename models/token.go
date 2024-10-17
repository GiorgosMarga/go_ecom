package models

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const ()

type RefreshToken struct {
	ID        primitive.ObjectID `bson:"_id"`
	UserId    primitive.ObjectID `bson:"user_id"`
	IsRevoked bool               `bson:"is_revoked"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

type TokenModel struct {
	coll *mongo.Collection
}

func NewRefreshToken(id primitive.ObjectID) RefreshToken {
	return RefreshToken{
		ID:        primitive.NewObjectID(),
		UserId:    id,
		IsRevoked: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (m TokenModel) Insert(rt *RefreshToken) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.coll.InsertOne(ctx, rt)
	if err != nil {
		return err
	}
	return nil
}

func (m TokenModel) GetForUser(tokenId string) (primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rt := RefreshToken{}
	id, err := primitive.ObjectIDFromHex(tokenId)
	if err != nil {
		return primitive.NilObjectID, ErrInvalidID
	}
	filter := bson.M{"_id": id, "is_revoked": false}
	err = m.coll.FindOne(ctx, filter).Decode(&rt)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return primitive.NilObjectID, ErrNotFound
		default:
			return primitive.NilObjectID, err
		}
	}
	return rt.UserId, nil
}
