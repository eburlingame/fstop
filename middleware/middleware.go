package middleware

import (
	"fmt"
	"net/http"

	. "github.com/eburlingame/fstop/resources"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const SESSION_USERNAME_KEY = "authed_user"

func IsLoggedIn(r *Resources, c *gin.Context) bool {
	session := sessions.Default(c)
	username := session.Get(SESSION_USERNAME_KEY)

	return username == r.Config.AdminUsername
}

func EnsureLoggedIn(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		if !IsLoggedIn(r, c) {
			session.Clear()
			session.Save()

			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		c.Set("isAdmin", true)
	}
}

func EnsureNotLoggedIn(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsLoggedIn(r, c) {
			c.Redirect(http.StatusFound, "/admin")
			c.Abort()
			return
		}
	}
}

func EnsureApiKeyPresent(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		correctHeader := fmt.Sprintf("Bearer %s", r.Config.ApiKey)
		headerValue := c.GetHeader("Authorization")

		if correctHeader != headerValue {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}
	}
}
