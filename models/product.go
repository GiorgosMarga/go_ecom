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

type Product struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Description string             `json:"description" bson:"description"`
	Price       float64            `json:"price" bson:"price"`
	Stock       int                `json:"-" bson:"stock"`
	Rating      int                `json:"rating" bason:"rating"`
	CreatedAt   time.Time          `json:"-" bson:"created_at"`
	UpdatedAt   time.Time          `json:"-" bson:"updated_at"`
}

type ProductModel struct {
	coll *mongo.Collection
}

type ProductUpdatePayload struct {
	Description *string  `json:"description" bson:"description"`
	Price       *float64 `json:"price" bson:"price"`
	Stock       *int     `json:"stock" bson:"stock"`
	Rating      *int     `json:"rating" bson:"rating"`
}

func validateDescription(v *validator.Validator, desc string) {
	v.Validate(validator.CheckLength(desc, 5, 5000), "description", "must be between 5 and 5000 characters long")
}
func validatePrice(v *validator.Validator, price float64) {
	v.Validate(price > 0, "price", "must be positive")
}
func validateStock(v *validator.Validator, stock int) {
	v.Validate(stock >= 0, "stock", "must be positive")
}
func ValidateProduct(v *validator.Validator, p Product) {
	validateDescription(v, p.Description)
	validatePrice(v, p.Price)
	validateStock(v, p.Stock)
}

func ValidateUpdatePayload(v *validator.Validator, p ProductUpdatePayload) {
	if p.Description != nil {
		validateDescription(v, *p.Description)
	}
	if p.Price != nil {
		validatePrice(v, *p.Price)
	}
	if p.Stock != nil {
		validateStock(v, *p.Stock)
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	p := &Product{}

	filter := bson.M{"_id": productID}
	if err := m.coll.FindOne(ctx, filter).Decode(&p); err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return p, nil
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

func (m ProductModel) GetPriceForOrder(idToQuantity map[primitive.ObjectID]int) (float64, error) {
	ids := make([]primitive.ObjectID, 0)
	for k := range idToQuantity {
		ids = append(ids, k)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	products := make([]Product, 0)
	filter := bson.M{"_id": bson.M{"$in": ids}}
	cursor, err := m.coll.Find(ctx, filter)
	if err != nil {
		return 0, err
	}

	if err := cursor.All(ctx, &products); err != nil {
		return 0, err
	}

	total := 0.0
	for _, product := range products {
		total += float64(idToQuantity[product.ID]) * product.Price
	}
	return total, err
}
