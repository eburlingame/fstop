package main

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AdminGetHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		username := session.Get("authed_user")
		if username == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		c.HTML(http.StatusOK, "admin.html", gin.H{
			"title":    "Dashboard",
			"username": username,
		})
	}
}
