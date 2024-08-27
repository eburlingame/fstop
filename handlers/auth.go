package handlers

import (
	"crypto/subtle"
	"fmt"
	"net/http"

	. "github.com/eburlingame/fstop/middleware"
	. "github.com/eburlingame/fstop/resources"
	"golang.org/x/crypto/bcrypt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func ViewerLoginGetHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get(VIEWER_SESSION_USERNAME_KEY)

		if username != nil {
			c.Redirect(http.StatusFound, "/")
			return
		}

		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "Login",
		})
	}
}

func AdminLoginGetHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get(ADMIN_SESSION_USERNAME_KEY)

		if username != nil {
			c.Redirect(http.StatusFound, "/admin")
			return
		}

		c.HTML(http.StatusOK, "admin_login.html", gin.H{
			"title": "Login",
		})
	}
}

func ViewerLoginPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		type LoginFormData struct {
			Password string `form:"password"`
		}

		var formData LoginFormData
		err := c.Bind(&formData)
		if err != nil {
			c.Status(400)
			return
		}

		redirect := c.DefaultQuery("redirect", "/")
		fmt.Printf("Redirect: %s", redirect)

		if comparePasswords(r.Config.ViewerPasswordHash, formData.Password) {
			session.Set(VIEWER_SESSION_USERNAME_KEY, VIEWER_USERNAME)
			session.Save()

			c.Redirect(http.StatusFound, redirect)
			return
		}

		c.HTML(http.StatusForbidden, "login.html", gin.H{
			"title": "Login",
			"error": "Incorrect username or password",
		})
	}
}

func comparePasswords(passwordHash []byte, inputtedPassword string) bool {
	inputtedPasswordBytes := []byte(inputtedPassword)

	passwordMatches := bcrypt.CompareHashAndPassword(passwordHash, inputtedPasswordBytes)
	return passwordMatches == nil
}

func compareUsernamePassword(r *Resources, inputtedUsername string, inputtedPassword string) bool {
	adminUsername := []byte(r.Config.AdminUsername)
	inputtedUsernameBytes := []byte(inputtedUsername)

	usernameMatches := subtle.ConstantTimeCompare(adminUsername, inputtedUsernameBytes)
	if usernameMatches != 1 {
		return false
	}

	return comparePasswords(r.Config.AdminPasswordHash, inputtedPassword)
}

func AdminLoginPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		type LoginFormData struct {
			Username string `form:"username"`
			Password string `form:"password"`
		}

		var formData LoginFormData
		c.Bind(&formData)

		if compareUsernamePassword(r, formData.Username, formData.Password) {
			session.Set(ADMIN_SESSION_USERNAME_KEY, r.Config.AdminUsername)
			session.Save()

			c.Redirect(http.StatusFound, "/admin")
			return
		}

		c.HTML(http.StatusForbidden, "admin_login.html", gin.H{
			"title": "Login",
			"error": "Incorrect username or password",
		})
	}
}

func AdminLogoutPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		session.Clear()
		session.Save()

		c.Redirect(http.StatusFound, "/login")
	}
}
