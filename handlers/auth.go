package handlers

import (
	"crypto/subtle"
	"net/http"

	. "github.com/eburlingame/fstop/middleware"
	. "github.com/eburlingame/fstop/resources"
	"golang.org/x/crypto/bcrypt"

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

func compareUsernamePassword(r *Resources, inputtedUsername string, inputtedPassword string) bool {
	adminUsername := []byte(r.Config.AdminUsername)
	adminPasswordHash := r.Config.AdminPasswordHash

	inputtedUsernameBytes := []byte(inputtedUsername)
	inputtedPasswordBytes := []byte(inputtedPassword)

	usernameMatches := subtle.ConstantTimeCompare(adminUsername, inputtedUsernameBytes)
	if usernameMatches != 1 {
		return false
	}

	passwordMatches := bcrypt.CompareHashAndPassword(adminPasswordHash, inputtedPasswordBytes)
	return passwordMatches == nil
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

		if compareUsernamePassword(r, formData.Username, formData.Password) {
			session.Set(SESSION_USERNAME_KEY, r.Config.AdminUsername)
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
