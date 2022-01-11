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

	r := &Resources{
		Config:  config,
		Storage: storage,
		Db:      db,
	}

	router := gin.Default()

	store := cookie.NewStore([]byte(config.Secret))
	router.Use(sessions.Sessions("fstop-session", store))

	router.LoadHTMLGlob("./templates/*")

	if err != nil {
		log.Fatal(err)
	}

	router.Static("/static/", "./static/")

	router.GET("/", HomeGetHandler(r))
	router.GET("/image/:imageId", ImageGetHandler(r))

	router.GET("/albums", AlbumsListGetHandler(r))
	router.GET("/album/:albumSlug", SingleAlbumGetHandler(r))

	router.GET("/login", EnsureNotLoggedIn(r), LoginGetHandler(r))
	router.POST("/login", EnsureNotLoggedIn(r), LoginPostHandler(r))
	router.POST("/logout", EnsureLoggedIn(r), LogoutPostHandler(r))

	router.GET("/admin", EnsureLoggedIn(r), AdminGetHandler(r))

	router.GET("/admin/albums", EnsureLoggedIn(r), AdminAlbumsGetHandler(r))
	router.GET("/admin/albums/:albumSlug", EnsureLoggedIn(r), AdminEditAlbumGetHandler(r))
	router.GET("/admin/albums/:albumSlug/add", EnsureLoggedIn(r), AdminAddPhotosGetHandler(r))
	router.POST("/admin/albums", EnsureLoggedIn(r), AdminAddAlbumPostHandler(r))
	router.POST("/admin/albums/:albumSlug/add", EnsureLoggedIn(r), AdminAddPhotosPostHandler(r))
	router.POST("/admin/albums/:albumSlug", EnsureLoggedIn(r), AdminEditAlbumPostHandler(r))
	router.POST("/admin/albums/:albumSlug/delete", EnsureLoggedIn(r), AdminDeleteAlbumPostHandler(r))
	router.POST("/admin/images/:imageId/delete", EnsureLoggedIn(r), AdminDeleteImagePostHandler(r))
	router.DELETE("/admin/albums/:albumSlug/:imageId", EnsureLoggedIn(r), AdminRemoveImageFromAlbumPostHandler(r))

	router.GET("/admin/upload", EnsureLoggedIn(r), AdminUploadGet(r))
	router.POST("/admin/upload/sign", EnsureLoggedIn(r), AdminUploadSignedUrlPostHandler(r))

	router.GET("/admin/import", EnsureLoggedIn(r), AdminImportGet(r))
	router.POST("/admin/import", EnsureLoggedIn(r), AdminImportPostHandler(r))
	router.GET("/admin/import/status/:batchId", EnsureLoggedIn(r), AdminImportStatusGetHandler(r))

	return router
}

func main() {
	r := setupRouter()
	// Listen and serve on 0.0.0.0:8080
	r.Run(":8080")
}
