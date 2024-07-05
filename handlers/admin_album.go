package handlers

import (
	"log"
	"net/http"

	. "github.com/eburlingame/fstop/models"
	. "github.com/eburlingame/fstop/resources"
	. "github.com/eburlingame/fstop/utils"

	"github.com/gin-gonic/gin"
)

func AdminAlbumsGetHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		var albums []Album
		r.Db.ListAlbums(&albums)

		c.HTML(http.StatusOK, "admin_albums.html", gin.H{
			"albums": albums,
		})
	}
}

func AdminAddAlbumPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		album := Album{
			AlbumId:     Uuid(),
			Slug:        "untitled",
			Name:        "Untitled",
			Description: "",
			IsPublished: false,
		}

		r.Db.AddAlbum(album)

		c.Redirect(http.StatusFound, "/admin/albums/"+album.Slug)
	}
}

type AlbumUriParams struct {
	AlbumSlug string `uri:"albumSlug" binding:"required"`
}

func EditAlbumPage(r *Resources, c *gin.Context) {
	var params AlbumUriParams

	err := c.BindUri(&params)
	if err != nil {
		c.Status(404)
		return
	}

	var album Album

	r.Db.GetAlbumBySlug(&album, params.AlbumSlug)
	files, _ := r.Db.ListAlbumImages(params.AlbumSlug, 400, 200, 0)

	c.HTML(http.StatusOK, "edit_album.html", gin.H{
		"album": album,
		"files": files,
	})
}

func AdminEditAlbumGetHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		EditAlbumPage(r, c)
	}
}

func AdminAddPhotosGetHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params AlbumUriParams

		err := c.BindUri(&params)
		if err != nil {
			c.Status(404)
			return
		}

		var album Album
		r.Db.GetAlbumBySlug(&album, params.AlbumSlug)

		files, err := r.Db.ListLatestFiles(400, 100, 0)

		c.HTML(200, "add_to_album.html", gin.H{
			"files": files,
			"album": album,
		})
	}
}

func AdminAddPhotosPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params AlbumUriParams

		err := c.BindUri(&params)
		if err != nil {
			c.Status(404)
			return
		}

		var album Album
		r.Db.GetAlbumBySlug(&album, params.AlbumSlug)

		imageIds := c.PostFormArray("images")

		for _, imageId := range imageIds {
			r.Db.AddImageToAlbum(album.AlbumId, imageId)
		}

		c.Redirect(http.StatusFound, "/admin/albums/"+album.Slug)
	}
}

func AdminEditAlbumPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params AlbumUriParams

		err := c.BindUri(&params)
		if err != nil {
			c.Status(404)
			return
		}

		type FormData struct {
			Name        string `form:"name"`
			Slug        string `form:"slug"`
			Description string `form:"description"`
			IsPublished string `form:"is_published"`
		}

		var form FormData
		c.Bind(&form)

		var album Album
		r.Db.GetAlbumBySlug(&album, params.AlbumSlug)

		slugChanged := album.Slug != form.Slug

		album.Name = form.Name
		album.Slug = form.Slug
		album.Description = form.Description
		album.IsPublished = form.IsPublished == "on"

		r.Db.UpdateAlbum(album.AlbumId, &album)

		if slugChanged {
			c.Redirect(http.StatusFound, "/admin/albums/"+album.Slug)
		} else {
			EditAlbumPage(r, c)
		}
	}
}

func AdminRemoveImageFromAlbumPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		type DeleteAlbumImageUriParams struct {
			AlbumSlug string `uri:"albumSlug" binding:"required"`
			ImageId   string `uri:"imageId" binding:"required"`
		}

		var params DeleteAlbumImageUriParams
		c.BindUri(&params)

		var album Album
		r.Db.GetAlbumBySlug(&album, params.AlbumSlug)
		r.Db.RemoveImageFromAlbum(album.AlbumId, params.ImageId)

		c.HTML(200, "image_removed.html", gin.H{})
	}
}

func deleteAndRemoveImage(r *Resources, imageId string) {
	// Remove images from storage
	var files []File
	r.Db.ListImageFiles(&files, imageId)

	for _, file := range files {
		err := r.Storage.DeleteFile(file.StoragePath)
		if err != nil {
			log.Printf("Error deleting file: %s\n", err)
		}
	}

	// Remove images from database
	r.Db.DeleteImage(imageId)
}

func AdminDeleteAlbumPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		type DeleteAlbumUriParams struct {
			AlbumSlug string `uri:"albumSlug" binding:"required"`
		}

		var params DeleteAlbumUriParams
		c.BindUri(&params)

		images, _ := r.Db.ListAlbumFiles(params.AlbumSlug)

		var album Album
		r.Db.GetAlbumBySlug(&album, params.AlbumSlug)

		for _, file := range images {
			println("Deleting image: ", file.ImageId)
			deleteAndRemoveImage(r, file.ImageId)
		}

		println(len(images), " images deleted from album")
		r.Db.DeleteAlbum(album.AlbumId)

		c.Redirect(http.StatusFound, "/admin/albums")
	}
}

func AdminDeleteImagePostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		type DeleteImageUriParams struct {
			ImageId string `uri:"imageId" binding:"required"`
		}

		var params DeleteImageUriParams
		c.BindUri(&params)

		deleteAndRemoveImage(r, params.ImageId)

		c.Redirect(http.StatusFound, "/")
	}
}
