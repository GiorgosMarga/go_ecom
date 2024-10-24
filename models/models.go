package models

import "go.mongodb.org/mongo-driver/v2/mongo"

type Models struct {
	User    UserModel
	Product ProductModel
	Cart    CartModel
	Token   TokenModel
	Order   OrderModel
}

func NewModels(db *mongo.Database) Models {
	return Models{
		User:    UserModel{coll: db.Collection("users", nil)},
		Product: ProductModel{coll: db.Collection("products", nil)},
		Token:   TokenModel{coll: db.Collection("tokens", nil)},
		Cart:    CartModel{coll: db.Collection("products", nil)},
		Order: OrderModel{
			productsColl: db.Collection("products", nil),
			orderColl:    db.Collection("orders", nil),
		},
	}
}
