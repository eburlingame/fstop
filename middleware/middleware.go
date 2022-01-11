package middleware

import (
	"net/http"

	. "github.com/eburlingame/fstop/resources"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const SESSION_USERNAME_KEY = "authed_user"

func EnsureLoggedIn(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get(SESSION_USERNAME_KEY)

		if username != r.Config.AdminUsername {
			session.Clear()
			session.Save()

			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
	}
}

func EnsureNotLoggedIn(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get(SESSION_USERNAME_KEY)

		if username == r.Config.AdminUsername {
			c.Redirect(http.StatusFound, "/admin")
			c.Abort()
			return
		}
	}
}
