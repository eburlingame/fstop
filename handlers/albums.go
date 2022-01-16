package handlers

import (
	"net/http"

	. "github.com/eburlingame/fstop/models"
	. "github.com/eburlingame/fstop/resources"

	"github.com/gin-gonic/gin"
)

func AlbumsListGetHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {

		albums, _ := r.Db.ListAlbumsCovers(true, 400, 100, 0)

		c.HTML(http.StatusOK, "albums.html", gin.H{
			"albums": albums,
		})
	}
}

func SingleAlbumGetHandler(r *Resources) gin.HandlerFunc {
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

		var album Album

		r.Db.GetAlbumBySlug(&album, params.AlbumSlug)
		files, _ := r.Db.ListAlbumImages(params.AlbumSlug, 500, 50, 0)

		c.HTML(http.StatusOK, "album.html", gin.H{
			"album": album,
			"files": files,
		})
	}
}
