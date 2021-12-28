package handlers

import (
	"net/http"

	. "github.com/eburlingame/fstop/resources"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func LoginGetHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("authed_user")

		if username != nil {
			c.Redirect(http.StatusFound, "/admin")
			return
		}

		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "Login",
		})
	}
}

func LoginPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		type LoginFormData struct {
			Username string `form:"username"`
			Password string `form:"password"`
		}

		var formData LoginFormData
		c.Bind(&formData)

		if formData.Username == "test" && formData.Password == "test" {
			session.Set("authed_user", "test")
			session.Save()

			c.Redirect(http.StatusFound, "/admin")
			return
		}

		c.HTML(http.StatusForbidden, "login.html", gin.H{
			"title": "Login",
			"error": "Incorrect username or password",
		})
	}
}

func LogoutPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		session.Clear()
		session.Save()

		c.Redirect(http.StatusFound, "/")
	}
}
