package models

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Review struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	ProductID primitive.ObjectID `json:"product_id" bson:"product_id"`
	Content   string             `json:"content" bson:"content"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}
type ReviewModel struct {
	coll *mongo.Collection
}

type ReviewPayload struct {
	ProductID primitive.ObjectID `json:"product_id"`
	Content   string             `json:"content"`
}
type ReviewUpdatePayload struct {
	Content *string `json:"content"`
}

func NewReview(userID, productID primitive.ObjectID, content string) *Review {
	return &Review{
		UserID:    userID,
		ProductID: productID,
		ID:        primitive.NewObjectID(),
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (m ReviewModel) Insert(review *Review) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.coll.InsertOne(ctx, review)
	if err != nil {
		return err
	}
	return nil
}
func (m ReviewModel) Delete(id primitive.ObjectID, user *UserInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var filter bson.M

	// an admin must be able to delete reviews from any user
	// but a user can delete only its own reviews
	if user.Role == GetRole(AdminRole) {
		filter = bson.M{"_id": id}
	} else {
		filter = bson.M{"user_id": user.UserID, "_id": id}
	}
	res, err := m.coll.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}
func (m ReviewModel) Get(id primitive.ObjectID) (*Review, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	review := &Review{}
	filter := bson.M{"_id": id}
	err := m.coll.FindOne(ctx, filter).Decode(&review)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return review, nil
}

func (m ReviewModel) Update(review *Review) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	filter := bson.M{"_id": review.ID}
	update := bson.D{{Key: "$set", Value: review}}
	res, err := m.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNotFound
	}
	return nil
}
