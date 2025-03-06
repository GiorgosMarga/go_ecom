// TODO: Price should be int and not float
// TODO: Order should just have product_id, variant_id, quantity and size

package models

import (
	"context"
	"errors"
	"time"

	"github.com/GiorgosMarga/ecom_go/internal/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	StatusPending = iota
	StatusPayed
	StatusShipped
	StatusDelivered
	StatusCanceled
)

type OrderProducts struct {
	Variant  primitive.ObjectID `json:"variant_id" bson:"variant_id"`
	Size     string             `json:"size" bson:"size"`
	Quantity int                `json:"quantity" bson:"quantity"`
}

type Order struct {
	ID              primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserId          primitive.ObjectID `json:"user_id" bson:"user_id"`
	Products        []OrderProducts    `json:"products" bson:"products"`
	Total           int                `json:"total" bson:"total"`
	Status          int                `json:"status" bson:"status"`
	PaymentIntentId string             `json:"payment_intent_id" bson:"payment_intent_id"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}

type OrderModel struct {
	coll *mongo.Collection
}

type OrderUpdatePayload struct {
	Products []OrderProducts `json:"products" bson:"products"`
	Status   *int            `json:"status" bson:"status"`
}

func ValidateOrderUpdatePayload(v *validator.Validator, payload OrderUpdatePayload) {
	if payload.Status != nil {
		v.Validate(validator.IsAllowedValue(*payload.Status, []int{StatusPending,
			StatusPayed,
			StatusShipped,
			StatusDelivered,
			StatusCanceled}), "status", "not allowed status")
	}
}

func (m OrderModel) Insert(order *Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	order.ID = primitive.NewObjectID()
	order.Status = StatusPending
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	_, err := m.coll.InsertOne(ctx, order)
	if err != nil {
		return err
	}
	return nil
}

func (m OrderModel) Get(id primitive.ObjectID) (*Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	order := &Order{}
	filter := bson.M{"_id": id}

	err := m.coll.FindOne(ctx, filter).Decode(&order)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return order, err
}

func (m OrderModel) Delete(id primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}

	res, err := m.coll.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (m OrderModel) Update(order *Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	filter := bson.M{"_id": order.ID}
	update := bson.D{{Key: "$set", Value: order}}

	res, err := m.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.ModifiedCount == 0 {
		return ErrNotFound
	}
	return nil
}
