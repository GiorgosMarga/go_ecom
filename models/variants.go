package models

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ProductVariant struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	ProductId primitive.ObjectID `json:"product_id" bson:"product_id"`
	Color     string             `json:"color" bson:"color"`
	Sizes     []int              `json:"size"`
	Stocks    []int              `json:"stock"`
	Img       []string           `json:"img" bson:"img"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

type VariantModel struct {
	coll      *mongo.Collection
	sizesColl *mongo.Collection
}

func (m VariantModel) insertSizesAndStocks(v ProductVariant) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fmt.Println(v.Sizes, v.Stocks)
	docs := make([]bson.M, len(v.Sizes))
	for idx := range v.Sizes {
		doc := bson.M{"_id": primitive.NewObjectID(), "variance_id": v.ID, "size": v.Sizes[idx], "stock": v.Stocks[idx]}
		docs[idx] = doc
	}
	// fmt.Println("I AM HERE", docs)
	_, err := m.sizesColl.InsertMany(ctx, docs)
	if err != nil {
		return err
	}
	return nil

}
func (m VariantModel) Insert(variant *ProductVariant) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	variant.ID = primitive.NewObjectID()
	variant.CreatedAt = time.Now()
	variant.UpdatedAt = time.Now()

	err := m.insertSizesAndStocks(*variant)
	if err != nil {
		return err
	}

	_, err = m.coll.InsertOne(ctx, variant)
	if err != nil {
		return err
	}
	return nil
}

func (m VariantModel) GetByProductId(productId string) ([]ProductVariant, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	variants := make([]ProductVariant, 0)
	objectId, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, ErrInvalidID
	}

	filter := bson.M{"product_id": objectId}

	cursor, err := m.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	if err := cursor.All(ctx, &variants); err != nil {
		return nil, err
	}
	return variants, nil
}

func (m VariantModel) Update(pv ProductVariant) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": pv.ID}
	update := bson.D{
		{Key: "$set", Value: pv},
	}

	res, err := m.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (m VariantModel) DeleteForProduct(product_id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(product_id)
	if err != nil {
		return ErrInvalidID
	}

	filter := bson.M{"product_id": objectId}

	res, err := m.coll.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil

}
func (m VariantModel) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidID
	}

	filter := bson.M{"_id": objectId}

	res, err := m.coll.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}
