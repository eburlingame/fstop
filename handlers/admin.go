package handlers

import (
	"fmt"
	"net/http"

	. "github.com/eburlingame/fstop/models"
	. "github.com/eburlingame/fstop/resources"

	"github.com/gin-gonic/gin"
)

func AdminGetHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {

		c.HTML(http.StatusOK, "admin.html", gin.H{
			"title": "Dashboard",
		})
	}
}

func AdminAlbumsGetHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {

		var albumCovers []AlbumFile
		r.Db.ListAlbumsCovers(&albumCovers, 400, 1000, 0)

		fmt.Printf("%#v", albumCovers)

		c.HTML(http.StatusOK, "admin_albums.html", gin.H{
			"albums": albumCovers,
		})
	}
}

func AdminEditAlbumGetHandler(r *Resources) gin.HandlerFunc {
	type UriParams struct {
		AlbumSlug string `uri:"albumSlug" binding:"required"`
	}

	return func(c *gin.Context) {
		var params UriParams

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
}
