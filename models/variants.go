// TODO: refactor variants

package models

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/GiorgosMarga/ecom_go/internal/validator"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var colors []string = []string{"red", "blue", "white", "black", "pink", "yellow", "gray"}

type SizesAndStock struct {
	Size  string `json:"size" bson:"size"`
	Stock int    `json:"stock" bson:"stock"`
}
type Variant struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	ProductId primitive.ObjectID `json:"product_id,omitempty" bson:"product_id,omitempty"`
	Color     string             `json:"color" bson:"color"`
	Sizes     []SizesAndStock    `json:"sizes,omitempty" bson:"sizes,omitempty"`
	Img       []string           `json:"img" bson:"img"`
	CreatedAt time.Time          `json:"-" bson:"created_at"`
	UpdatedAt time.Time          `json:"-" bson:"updated_at"`
}

type VariantModel struct {
	coll     *mongo.Collection
	infoColl *mongo.Collection
}

func validateSizesInfo(v *validator.Validator, sizesInfo []SizesAndStock) {
	for _, info := range sizesInfo {
		v.Validate(len(info.Size) > 0, "size", "cant be empty")
		v.Validate(info.Stock >= 0, "stock", "cant be negative")
	}
}
func validateColor(v *validator.Validator, pv Variant) {
	v.Validate(slices.Contains(colors, pv.Color), "color", "unknown color")
}

func ValidateVariant(v *validator.Validator, pv Variant) {
	validateColor(v, pv)
	validateSizesInfo(v, pv.Sizes)
}

func (m VariantModel) Insert(variant *Variant) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	variant.ID = primitive.NewObjectID()
	variant.CreatedAt = time.Now()
	variant.UpdatedAt = time.Now()

	_, err := m.coll.InsertOne(ctx, variant)
	if err != nil {
		return err
	}
	return nil
}

func (m VariantModel) GetByProductId(productId string) ([]Variant, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	variants := make([]Variant, 0)
	objectId, err := primitive.ObjectIDFromHex(productId)
	if err != nil {
		return nil, ErrInvalidID
	}

	filter := bson.M{"product_id": objectId}
	projection := bson.M{"product_id": 0}

	cursor, err := m.coll.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	if err := cursor.All(ctx, &variants); err != nil {
		return nil, err
	}
	return variants, nil
}

func (m VariantModel) GetById(id string) (*Variant, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	variantId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidID
	}

	filter := bson.M{"_id": variantId}
	var v Variant
	err = m.coll.FindOne(ctx, filter).Decode(&v)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return &v, nil
}

func (m VariantModel) Update(pv Variant) error {
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

	filter = bson.M{"variant_id": objectId}
	res, err = m.infoColl.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (m VariantModel) GetTotalPrice(order *Order) (int, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	variantsId := make([]primitive.ObjectID, 0)

	for _, p := range order.Products {
		variantsId = append(variantsId, p.Variant)
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"_id": bson.M{"$in": variantsId}}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "products",
			"localField":   "product_id",
			"foreignField": "_id",
			"as":           "product",
		}}},
		{{Key: "$unwind", Value: bson.M{"path": "$product"}}},
		{{Key: "$project", Value: bson.M{
			"price":      "$product.price", // Get price from the product
			"variant_id": "$_id",           // Keep variant ID
			"sizes":      "$sizes",         // Include sizes
		}}},
	}

	cursor, err := m.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}

	defer cursor.Close(ctx)

	type tempProduct struct {
		Price     int                `bson:"price"`
		VariantId primitive.ObjectID `bson:"variant_id"`
		Sizes     []SizesAndStock    `bson:"sizes"`
	}

	var products []tempProduct

	if err := cursor.All(ctx, &products); err != nil {
		return 0, err
	}

	if err := cursor.Err(); err != nil {
		return 0, err
	}
	fmt.Println(products)
	var total int = 0

	for _, orderProducts := range order.Products {
		for _, product := range products {
			if product.VariantId == orderProducts.Variant {
				for _, size := range product.Sizes {
					if size.Size == orderProducts.Size && size.Stock >= orderProducts.Quantity {
						total += size.Stock * product.Price * orderProducts.Quantity
					}
				}
			}
		}
	}

	if total == 0 {
		return 0, errors.New("invalid order")
	}
	return total, nil
}
