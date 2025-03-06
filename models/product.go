package models

import (
	"context"
	"fmt"
	"time"

	"github.com/GiorgosMarga/ecom_go/internal/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Product struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Tags        string             `json:"tags" bson:"tags"`
	Description string             `json:"description" bson:"description"`
	Price       int                `json:"price" bson:"price"`

	Variants  []Variant `json:"variants,omitempty" bson:"variants,omitmepty"`
	CreatedAt time.Time `json:"-" bson:"created_at"`
	UpdatedAt time.Time `json:"-" bson:"updated_at"`
}

type ProductModel struct {
	coll *mongo.Collection
}

type ProductUpdatePayload struct {
	Description *string `json:"description" bson:"description"`
	Price       *int    `json:"price" bson:"price"`
	Name        *string `json:"name" bson:"name"`
	Tags        *string `json:"tags" bson:"tags"`
}

func validateDescription(v *validator.Validator, desc string) {
	v.Validate(validator.CheckLength(desc, 5, 5000), "description", "must be between 5 and 5000 characters long")
}
func validatePrice(v *validator.Validator, price int) {
	v.Validate(price > 0, "price", "must be positive")
}

func validateImg(v *validator.Validator, img string) {
	v.Validate(len(img) > 0, "img", "must be provided")
}
func validateName(v *validator.Validator, name string) {
	v.Validate(len(name) > 0, "name", "must be provided")
}

func validateTags(v *validator.Validator, tags string) {
	v.Validate(len(tags) >= 0, "tags", "must be provided")
}
func ValidateProduct(v *validator.Validator, p Product) {
	validateDescription(v, p.Description)
	validatePrice(v, p.Price)
	// validateImg(v, p.Img)
	validateName(v, p.Name)
	validateTags(v, p.Tags)
}

func (m ProductModel) Insert(p *Product) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	p.ID = primitive.NewObjectID()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	_, err := m.coll.InsertOne(ctx, p)
	if err != nil {
		return err
	}
	return nil
}

func (m ProductModel) GetById(id string) (*Product, error) {

	productID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidID
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	p := Product{}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"_id": productID}}}, // Filter by specific product ID
		{{Key: "$lookup", Value: bson.M{
			"from":         "variants",
			"localField":   "_id",
			"foreignField": "product_id",
			"as":           "variants",
		}}},
		{{Key: "$unwind", Value: bson.M{"path": "$variants", "preserveNullAndEmptyArrays": true}}},
		{{Key: "$group", Value: bson.M{
			"_id":         "$_id",
			"name":        bson.D{{Key: "$first", Value: "$name"}},
			"tags":        bson.D{{Key: "$first", Value: "$tags"}},
			"description": bson.D{{Key: "$first", Value: "$description"}},
			"price":       bson.D{{Key: "$first", Value: "$price"}},
			"variants":    bson.D{{Key: "$push", Value: "$variants"}},
			"created_at":  bson.D{{Key: "$first", Value: "$created_at"}},
			"updated_at":  bson.D{{Key: "$first", Value: "$updated_at"}},
		}}},
	}
	cursor, err := m.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return nil, ErrNotFound
	}
	if err := cursor.Decode(&p); err != nil {
		return nil, err
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	fmt.Println(p)

	return &p, nil
}

func (m ProductModel) Delete(id string) error {
	productID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	filter := bson.M{"_id": productID}
	res, err := m.coll.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return ErrNotFound
	}

	return nil
}

func (m ProductModel) Update(product *Product) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	product.UpdatedAt = time.Now()
	filter := bson.M{"_id": product.ID}
	update := bson.D{{Key: "$set", Value: product}}
	res, err := m.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}
