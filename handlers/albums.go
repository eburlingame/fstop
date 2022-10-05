package handlers

import (
	"fmt"
	"net/http"

	. "github.com/eburlingame/fstop/models"
	. "github.com/eburlingame/fstop/resources"
	. "github.com/eburlingame/fstop/utils"

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
		var imagesWithSrcSets []ImageWithSrcSet

		r.Db.GetAlbumBySlug(&album, params.AlbumSlug)
		images, _ := r.Db.ListAlbumFiles(params.AlbumSlug)

		for _, img := range images {
			metaDescription := fmt.Sprintf("%s, %s (%s' f/%.1f ISO %.0f)", img.CameraModel, img.Lens, img.ShutterSpeed, img.FNumber, img.ISO)

			imagesWithSrcSets = append(imagesWithSrcSets, ImageWithSrcSet{
				ImageId:       img.ImageId,
				SrcSet:        ComputeImageSrcSet((img.Files)),
				SmallImageUrl: FindSizedImage(img.Files, 500).PublicURL,
				Width:         img.Width,
				Height:        img.Height,
				Title:         img.DateTimeOriginal.Format("Monday, January _2, 2006"),
				Description:   metaDescription,
			})
		}

		c.HTML(http.StatusOK, "album.html", gin.H{
			"album":  album,
			"images": imagesWithSrcSets,
		})
	}
}
