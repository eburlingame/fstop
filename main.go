package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.LoadHTMLGlob("./templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{
			"title": "Main website",
		})
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "Login",
		})
	})

	r.POST("/login", func(c *gin.Context) {
		type LoginFormData struct {
			Username string `form: "username"`
			Password string `form: "password"`
		}

		var formData LoginFormData

		fmt.Println(c.PostForm("username"))
		fmt.Println(c.PostForm("password"))

		c.ShouldBind(&formData)
		fmt.Println(formData)

		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "Login",
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
