package main

import (
	"log"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

type Resources struct {
	config  *Configuration
	storage Storage
}

func setupRouter() *gin.Engine {
	config := GetConfig()

	// Create a single AWS session (we can re use this if we're uploading many files)
	storage, err := InitS3Storage(config)

	resources := &Resources{
		config:  config,
		storage: storage,
	}

	r := gin.Default()

	// Set a lower memory limit for multipart forms (default is 32 MiB)
	r.MaxMultipartMemory = 128 << 20 // 128 MiB

	store := cookie.NewStore([]byte("mysecret"))
	r.Use(sessions.Sessions("mysession", store))

	r.LoadHTMLGlob("./templates/*")

	if err != nil {
		log.Fatal(err)
	}

	r.GET("/", HomeGetHandler(resources))

	r.GET("/login", LoginGetHandler(resources))
	r.POST("/login", LoginPostHandler(resources))

	r.GET("/admin", AdminGetHandler(resources))
	r.GET("/admin/upload", AdminUploadGet(resources))
	r.POST("/admin/upload", UploadHandler(resources))

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
