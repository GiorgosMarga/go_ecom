package main

import (
	"errors"
	"time"

	"github.com/GiorgosMarga/ecom_go/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrWrongHeader = errors.New("bad header: 'content-type'")
	ErrInvalidJWT  = errors.New("invalid token")
)

func WriteJSON(c *gin.Context, status int, v any) {
	c.JSON(status, v)
}

func ReadJSON(c *gin.Context, dst any) error {
	header := c.Request.Header.Get("Content-Type")
	if header != "application/json" {
		return ErrWrongHeader
	}
	return c.Bind(&dst)
}

// creates the access and refresh token
func (app *application) createAccessToken(user models.User) (string, error) {

	registeredClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "goecom",
		Audience:  jwt.ClaimStrings{user.ID.Hex()},
	}

	claims := models.UserTokenClaims{
		UserInfo: models.UserInfo{
			UserID: user.ID,
			Email:  user.Email,
			Role:   user.Role,
		},
		RegisteredClaims: registeredClaims,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(app.cfg.jwtSecret)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}
func (app *application) createRefreshToken(user models.User) (string, error) {
	rt := models.RefreshToken{
		ID:        primitive.NewObjectID(),
		UserId:    user.ID,
		IsRevoked: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := app.models.Token.Insert(&rt); err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 7)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "goecom",
		Audience:  jwt.ClaimStrings{user.ID.Hex()},
		Subject:   rt.ID.Hex(),
	})
	signedToken, err := token.SignedString(app.cfg.jwtSecret)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}
func (app *application) verifyToken(token string) (*jwt.Token, error) {
	t, err := jwt.ParseWithClaims(token, &models.UserTokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		return app.cfg.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, ErrInvalidJWT
	}

	return t, nil
}

func GetUser(c *gin.Context) (*models.UserInfo, error) {
	val, ok := c.Get("user")
	if !ok {
		return nil, ErrInvalidJWT
	}
	user := val.(models.UserInfo)
	return &user, nil
}

func ReadIdParam(c *gin.Context) primitive.ObjectID {
	val := c.Param("id")
	if val == "" {
		return primitive.NilObjectID
	}
	id, err := primitive.ObjectIDFromHex(val)
	if err != nil {
		return primitive.NilObjectID
	}
	return id
}
