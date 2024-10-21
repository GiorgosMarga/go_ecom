package models

import (
	"context"
	"time"

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
	ProductId primitive.ObjectID `json:"id" bson:"id"`
	Quantity  int                `json:"quantity" bson:"quantity"`
}

type Order struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	UserId    primitive.ObjectID `json:"user_id" bson:"user_id"`
	Products  []OrderProducts    `json:"products" bson:"products"`
	Total     float64            `json:"total" bson:"total"`
	Status    int                `json:"status" bson:"status"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

type OrderModel struct {
	orderColl    *mongo.Collection
	productsColl *mongo.Collection
}

func getProductsIds(products []OrderProducts) []primitive.ObjectID {
	ids := make([]primitive.ObjectID, len(products))
	for i, product := range products {
		ids[i] = product.ProductId
	}
	return ids
}
func calculateOrderTotal(products []Product, order *Order) float64 {
	productsToQuantity := make(map[primitive.ObjectID]int)
	for _, p := range order.Products {
		productsToQuantity[p.ProductId] += p.Quantity
	}

	total := 0.0
	for _, p := range products {
		total += p.Price * float64(productsToQuantity[p.ID])
	}
	return total
}
func (m OrderModel) Insert(order *Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ids := getProductsIds(order.Products)
	filter := bson.M{"_id": bson.M{"$in": ids}}
	cursor, err := m.productsColl.Find(ctx, filter)
	if err != nil {
		return err
	}
	var products []Product
	if err := cursor.All(ctx, &products); err != nil {
		return err
	}

	order.ID = primitive.NewObjectID()
	order.Status = StatusPending
	order.Total = calculateOrderTotal(products, order)
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	_, err = m.orderColl.InsertOne(ctx, order)
	if err != nil {
		return err
	}
	return nil
}
