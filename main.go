package main

import (
	"io"
	"log"
	"os"

	. "github.com/eburlingame/fstop/handlers"
	. "github.com/eburlingame/fstop/middleware"

	. "github.com/eburlingame/fstop/process"
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

	queue, err := InitQueue(db.Db)
	if err != nil {
		log.Fatal(err)
	}

	r := &Resources{
		Config:  config,
		Storage: storage,
		Db:      db,
		Queue:   &queue,
	}

	go InitWorkers(r)

	gin.DisableConsoleColor()
	f, _ := os.OpenFile("fstop.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	log.SetOutput(gin.DefaultWriter)

	log.Println("Starting server")
	router := gin.Default()

	store := cookie.NewStore([]byte(config.Secret))
	router.Use(sessions.Sessions("fstop-session", store))

	router.LoadHTMLGlob("./templates/*")

	if err != nil {
		log.Fatal(err)
	}

	router.Static("/static/", "./static/")
	router.StaticFile("/favicon.ico", "./static/favicon.ico")
	router.StaticFile("/robots.txt", "./static/robots.txt")

	router.GET("/", EnsureLoggedIn(r), HomeGetHandler(r))
	router.GET("/image/:imageId", EnsureLoggedIn(r), ImageGetHandler(r))

	router.GET("/albums", EnsureLoggedIn(r), AlbumsListGetHandler(r))
	router.GET("/album/:albumSlug", EnsureLoggedIn(r), SingleAlbumGetHandler(r))

	router.GET("/login", EnsureNotLoggedIn(r), ViewerLoginGetHandler(r))
	router.POST("/login", EnsureNotLoggedIn(r), ViewerLoginPostHandler(r))
	router.POST("/logout", EnsureLoggedIn(r), AdminLogoutPostHandler(r))

	router.GET("/admin", EnsureAdminLoggedIn(r), AdminGetHandler(r))

	router.GET("/admin/albums", EnsureAdminLoggedIn(r), AdminAlbumsGetHandler(r))
	router.GET("/admin/albums/:albumSlug", EnsureAdminLoggedIn(r), AdminEditAlbumGetHandler(r))
	router.GET("/admin/albums/:albumSlug/add", EnsureAdminLoggedIn(r), AdminAddPhotosGetHandler(r))
	router.POST("/admin/albums", EnsureAdminLoggedIn(r), AdminAddAlbumPostHandler(r))
	router.POST("/admin/albums/:albumSlug/add", EnsureAdminLoggedIn(r), AdminAddPhotosPostHandler(r))
	router.POST("/admin/albums/:albumSlug", EnsureAdminLoggedIn(r), AdminEditAlbumPostHandler(r))
	router.POST("/admin/albums/:albumSlug/delete", EnsureAdminLoggedIn(r), AdminDeleteAlbumPostHandler(r))
	router.POST("/admin/images/:imageId/delete", EnsureAdminLoggedIn(r), AdminDeleteImagePostHandler(r))
	router.DELETE("/admin/albums/:albumSlug/:imageId", EnsureAdminLoggedIn(r), AdminRemoveImageFromAlbumPostHandler(r))

	router.GET("/admin/login", EnsureNotLoggedIn(r), AdminLoginGetHandler(r))
	router.POST("/admin/login", EnsureNotLoggedIn(r), AdminLoginPostHandler(r))
	router.POST("/admin/logout", EnsureAdminLoggedIn(r), AdminLogoutPostHandler(r))
	router.GET("/admin/upload", EnsureAdminLoggedIn(r), AdminUploadGet(r))
	router.POST("/admin/upload/sign", EnsureAdminLoggedIn(r), AdminUploadSignedUrlPostHandler(r))

	router.GET("/admin/import", EnsureAdminLoggedIn(r), AdminImportGet(r))
	router.POST("/admin/import", EnsureAdminLoggedIn(r), AdminImportPostHandler(r))
	router.GET("/admin/import/status/:batchId", EnsureAdminLoggedIn(r), AdminImportStatusGetHandler(r))

	router.POST("/api/v1/admin/import", EnsureApiKeyPresent(r), ImportApiPostHandler(r))
	router.POST("/api/v1/admin/resize/single", EnsureApiKeyPresent(r), SingleResizeApiPostHandler(r))
	router.POST("/api/v1/admin/resize", EnsureApiKeyPresent(r), BulkResizeApiPostHandler(r))
	router.POST("/api/v1/admin/purge", EnsureApiKeyPresent(r), PurgeOrphanImagesApiPostHandler(r))
	router.GET("/api/v1/admin/import/:batchId", EnsureApiKeyPresent(r), ImportStateApiGetHandler(r))

	return router
}

func main() {
	r := setupRouter()
	// Listen and serve on 0.0.0.0:8080
	r.Run(":8080")
}
