package models

import (
	"errors"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrUsedEmail = errors.New("email already in use")
	ErrInvalidID = errors.New("invalid id")
	ErrNotFound  = errors.New("resource doesn't exist")
)

type Models struct {
	User    UserModel
	Product ProductModel
	Cart    CartModel
	Token   TokenModel
	Order   OrderModel
	Review  ReviewModel
	Variant VariantModel
}

func NewModels(db *mongo.Database) Models {
	return Models{
		User:    UserModel{coll: db.Collection("users", nil)},
		Product: ProductModel{coll: db.Collection("products", nil), variantsColl: db.Collection("variants", nil)},
		Token:   TokenModel{coll: db.Collection("tokens", nil)},
		Cart:    CartModel{coll: db.Collection("products", nil)},
		Order: OrderModel{
			productsColl: db.Collection("products", nil), // no need
			orderColl:    db.Collection("orders", nil),
		},
		Review:  ReviewModel{coll: db.Collection("review", nil)},
		Variant: VariantModel{coll: db.Collection("variants", nil), sizesColl: db.Collection("sizes", nil)},
	}
}
