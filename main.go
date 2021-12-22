package main

import (
	"log"

	. "github.com/eburlingame/fstop/handlers"
	. "github.com/eburlingame/fstop/middleware"
	. "github.com/eburlingame/fstop/resources"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

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
		Config:  config,
		Storage: storage,
		Db:      db,
	}

	r := gin.Default()

	store := cookie.NewStore([]byte("mysecret"))
	r.Use(sessions.Sessions("mysession", store))

	r.LoadHTMLGlob("./templates/*")

	if err != nil {
		log.Fatal(err)
	}

	r.StaticFile("/static/style.css", "./static/style.css")

	r.GET("/", HomeGetHandler(resources))
	r.GET("/image/:imageId", ImageGetHandler(resources))

	r.GET("/albums", AlbumsListGetHandler(resources))
	r.GET("/album/:albumSlug", SingleAlbumGetHandler(resources))

	r.GET("/login", EnsureNotLoggedIn(), LoginGetHandler(resources))
	r.POST("/login", EnsureNotLoggedIn(), LoginPostHandler(resources))

	r.GET("/admin", EnsureLoggedIn(), AdminGetHandler(resources))

	r.GET("/admin/albums", EnsureLoggedIn(), AdminAlbumsGetHandler(resources))
	r.GET("/admin/albums/:albumSlug", EnsureLoggedIn(), AdminEditAlbumGetHandler(resources))
	r.GET("/admin/albums/:albumSlug/add", EnsureLoggedIn(), AdminAddPhotosGetHandler(resources))
	r.POST("/admin/albums", EnsureLoggedIn(), AdminAddAlbumPostHandler(resources))
	r.POST("/admin/albums/:albumSlug/add", EnsureLoggedIn(), AdminAddPhotosPostHandler(resources))
	r.POST("/admin/albums/:albumSlug", EnsureLoggedIn(), AdminEditAlbumPostHandler(resources))
	r.POST("/admin/albums/:albumSlug/delete", EnsureLoggedIn(), AdminDeleteAlbumPostHandler(resources))
	r.DELETE("/admin/albums/:albumSlug/:imageId", EnsureLoggedIn(), AdminRemoveImageFromAlbumPostHandler(resources))

	r.GET("/admin/upload", EnsureLoggedIn(), AdminUploadGet(resources))
	r.POST("/admin/upload/sign", EnsureLoggedIn(), AdminUploadSignedUrlPostHandler(resources))

	r.GET("/admin/import", EnsureLoggedIn(), AdminImportGet(resources))
	r.POST("/admin/import", EnsureLoggedIn(), AdminImportPostHandler(resources))
	r.GET("/admin/import/status/:batchId", EnsureLoggedIn(), AdminImportStatusGetHandler(resources))

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
