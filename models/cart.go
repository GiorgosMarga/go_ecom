package models

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type CartProduct struct {
	Product
	Quantity int `json:"quantity" bson:"quantity"`
}

type Cart struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserId    primitive.ObjectID `json:"-" bson:"user_id"`
	Products  []CartProduct      `json:"products" bson:"products"`
	Total     float64            `json:"total" bson:"total"`
	Active    int                `json:"-" bson:"active"`
	CreatedAt time.Time          `json:"-" bson:"created_at"`
	UpdatedAt time.Time          `json:"-" bson:"updated_at"`
}

type CartModel struct {
	coll        *mongo.Collection
	productColl *mongo.Collection
}

type CartPayload struct {
	Products []primitive.ObjectID `json:"products" bson:"products"`
}

func getIds(products []CartProduct) []primitive.ObjectID {
	ids := make([]primitive.ObjectID, len(products))
	for i, product := range products {
		ids[i] = product.ID
	}
	return ids
}
func (m CartModel) Insert(c *Cart) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	c.ID = primitive.NewObjectID()
	c.Active = 1
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	_, err := m.coll.InsertOne(ctx, c)
	return err
}

func (m CartModel) Get(id primitive.ObjectID) (*Cart, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	c := &Cart{}
	filter := bson.M{"_id": id}
	err := m.coll.FindOne(ctx, filter).Decode(&c)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	filter = bson.M{"_id": bson.M{"$in": getIds(c.Products)}}

	cursor, err := m.productColl.Find(ctx, filter)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &c.Products); err != nil {
		return nil, err
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return c, nil
}

func (m CartModel) Update(c *Cart) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	c.UpdatedAt = time.Now()
	filter := bson.M{"_id": c.ID}
	update := bson.D{{Key: "$set", Value: c}}
	res, err := m.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (m CartModel) Delete(id, userId primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	filter := bson.M{"_id": id, "user_id": userId}
	res, err := m.coll.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}
