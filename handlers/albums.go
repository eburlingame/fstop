package handlers

import (
	"fmt"
	"net/http"

	. "github.com/eburlingame/fstop/models"
	. "github.com/eburlingame/fstop/resources"

	"github.com/gin-gonic/gin"
)

func AlbumsListGetHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		var files []AlbumFile

		r.Db.ListAlbumsCovers(&files, 400, 20, 0)

		fmt.Printf("PublicURL: %#v", files)

		c.HTML(http.StatusOK, "albums.html", gin.H{
			"files": files,
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
		var files []File

		r.Db.GetAlbumBySlug(&album, params.AlbumSlug)
		r.Db.ListAlbumImages(&files, params.AlbumSlug, 500, 50, 0)

		c.HTML(http.StatusOK, "album.html", gin.H{
			"album": album,
			"files": files,
		})
	}
}
