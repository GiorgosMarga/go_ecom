package main

import (
	"strings"

	"github.com/GiorgosMarga/ecom_go/models"
	"github.com/gin-gonic/gin"
)

func (app *application) authenticateUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		header := c.Request.Header["Authorization"][0]
		if header == "" {
			app.notAuthenticatedError(c)
			c.Abort()
			return
		}
		splittedHeader := strings.Split(header, " ")
		if len(splittedHeader) != 2 {
			app.notAuthenticatedError(c)
			c.Abort()
			return
		}

		accessToken := splittedHeader[1]

		jwtToken, err := app.verifyToken(accessToken)
		if err != nil {
			app.logger.Println(err.Error())
			app.notAuthenticatedError(c)
			c.Abort()
			return
		}

		claims, ok := jwtToken.Claims.(*models.UserTokenClaims)
		if !ok {
			app.notAuthenticatedError(c)
			c.Abort()
			return
		}
		c.Set("user", claims.UserInfo)
		c.Next()
	}
}

func (app *application) authorizeUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		val, ok := c.Get("user")
		if !ok {
			app.internalServerError(c, ErrInvalidJWT)
			c.Abort()
			return
		}
		user := val.(models.UserInfo)

		if user.Role != models.GetRole(models.AdminRole) {
			app.notAuthorizedError(c)
			c.Abort()
			return
		}
		c.Next()
	}
}
