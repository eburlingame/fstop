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
	db      Database
}

func setupRouter() *gin.Engine {
	config := GetConfig()

	storage, err := InitS3Storage(config)
	if err != nil {
		log.Fatal(err)
	}

	db, err := InitSqliteDatabase(config)
	if err != nil {
		log.Fatal(err)
	}

	resources := &Resources{
		config:  config,
		storage: storage,
		db:      db,
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

	r.GET("/login", ensureNotLoggedIn(), LoginGetHandler(resources))
	r.POST("/login", ensureNotLoggedIn(), LoginPostHandler(resources))

	r.GET("/admin", ensureLoggedIn(), AdminGetHandler(resources))
	r.GET("/admin/upload", ensureLoggedIn(), AdminUploadGet(resources))
	r.POST("/admin/upload", ensureLoggedIn(), AdminUploadPostHandler(resources))

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
