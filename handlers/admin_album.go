package handlers

import (
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

	var files []File
	var album Album

	r.Db.GetAlbumBySlug(&album, params.AlbumSlug)
	r.Db.ListAlbumImages(&files, params.AlbumSlug, 400, 1000, 0)

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

		var files []File
		r.Db.ListLatestPhotos(&files, 200, 100, 0)

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

func AdminDeleteAlbumPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		type DeleteAlbumUriParams struct {
			AlbumSlug string `uri:"albumSlug" binding:"required"`
		}

		var params DeleteAlbumUriParams
		c.BindUri(&params)

		var album Album
		r.Db.GetAlbumBySlug(&album, params.AlbumSlug)
		r.Db.DeleteAlbum(album.AlbumId)

		c.Redirect(http.StatusFound, "/admin/albums")
	}
}
