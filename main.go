package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	store := cookie.NewStore([]byte("mysecret"))
	r.Use(sessions.Sessions("mysession", store))

	r.LoadHTMLGlob("./templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "Main website",
		})
	})

	r.GET("/login", func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("authed_user")

		if username != nil {
			c.Redirect(http.StatusFound, "/admin")
			return
		}

		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "Login",
		})
	})

	r.POST("/login", func(c *gin.Context) {
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
	})

	r.GET("/admin", func(c *gin.Context) {
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
	})

	r.GET("/admin/upload", func(c *gin.Context) {
		session := sessions.Default(c)

		username := session.Get("authed_user")
		if username == nil {
			c.Redirect(http.StatusFound, "/login")
			return
		}

		c.HTML(http.StatusOK, "upload.html", gin.H{
			"title":    "Dashboard",
			"username": username,
		})
	})

	// Set a lower memory limit for multipart forms (default is 32 MiB)
	r.MaxMultipartMemory = 128 << 20 // 128 MiB

	r.POST("/upload", func(c *gin.Context) {
		// Multipart form
		form, _ := c.MultipartForm()
		files := form.File["upload"]

		for _, file := range files {
			log.Println(file.Filename)

			// Upload the file to specific dst.
			c.SaveUploadedFile(file, "./"+file.Filename)
		}
		c.String(http.StatusOK, fmt.Sprintf("%d files uploaded!", len(files)))
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
