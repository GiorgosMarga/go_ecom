package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/GiorgosMarga/ecom_go/internal/validator"
	"github.com/GiorgosMarga/ecom_go/utils"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)



type Role string

const (
	AdminRole = iota
	UserRole
)

func GetRole(role int) Role {
	switch role {
	case AdminRole:
		return "admin"
	case UserRole:
		return "user"
	default:
		return ""
	}
}

type UserModel struct {
	coll *mongo.Collection
}

type User struct {
	ID           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Email        string             `json:"email" bson:"email"`
	PasswordHash string             `json:"-" bson:"password_hash"`
	Role         Role               `json:"-" bson:"role"`
	CreatedAt    time.Time          `json:"-" bson:"created_at"`
	UpdatedAt    time.Time          `json:"-" bson:"updated_at"`
}

type UserPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type UserUpdatePayload struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
	Role     *Role   `json:"role"`
}
type UserTokenClaims struct {
	jwt.RegisteredClaims
	UserInfo
}

type UserInfo struct {
	Email  string
	UserID primitive.ObjectID
	Role   Role
}

func (role Role) Validate() bool {
	return role == GetRole(AdminRole) || role == GetRole(UserRole)
}
func validatePassword(v *validator.Validator, password string) {
	v.Validate(len(password) != 0, "password", "must be provided")
	v.Validate(validator.CheckLength(password, 6, 100), "password", "length must be between 6 and 100 characters")
}
func validateEmail(v *validator.Validator, email string) {
	v.Validate(len(email) != 0, "email", "must be provided")
	v.Validate(validator.CheckLength(email, 3, 100), "email", "length must be between 3 and 100 characters")
}
func ValidateUser(v *validator.Validator, u User) {
	validatePassword(v, u.PasswordHash)
	validateEmail(v, u.Email)
	if u.Role != "" {
		v.Validate(u.Role.Validate(), "role", "invalid")
	}
}

func (m UserModel) Insert(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	count, err := m.coll.CountDocuments(ctx, bson.M{"email": user.Email}, nil)
	if err != nil {
		return err
	}

	if count > 0 {
		return ErrUsedEmail
	}
	hashedPassword, err := utils.HashPassword(user.PasswordHash)
	if err != nil {
		return err
	}
	user.PasswordHash = hashedPassword
	_, err = m.coll.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

func (m UserModel) GetByID(id string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, ErrInvalidID
	}
	var u *User
	filter := bson.M{"_id": bson.M{"$eq": objectId}}
	if err := m.coll.FindOne(ctx, filter).Decode(&u); err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return u, nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var u *User
	if err := m.coll.FindOne(ctx, bson.M{"email": email}).Decode(&u); err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}
	return u, nil
}

func (m UserModel) Update(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	user.UpdatedAt = time.Now()
	update := bson.D{
		{Key: "$set", Value: user},
	}
	_, err := m.coll.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
