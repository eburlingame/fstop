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
		var files []Alb

		r.Db.ListAlbumsCovers(&files, 400, 20, 0)

		fmt.Printf("PublicURL: %#v", files)

		c.HTML(http.StatusOK, "albums.html", gin.H{
			"files": files,
		})
	}
}

func SingleAlbumGetHandler(r *Resources) gin.HandlerFunc {
	type UriParams struct {
		ImageId string `uri:"albumId" binding:"required"`
	}

	return func(c *gin.Context) {
		var params UriParams

		err := c.BindUri(&params)
		if err != nil {
			c.Status(404)
			return
		}

		var file File
		r.Db.GetFile(&file, params.ImageId, 1000)

		c.HTML(http.StatusOK, "image.html", gin.H{
			"file": file,
		})
	}
}
