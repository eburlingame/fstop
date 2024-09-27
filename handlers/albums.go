package handlers

import (
	"log"
	"net/http"

	. "github.com/eburlingame/fstop/models"
	. "github.com/eburlingame/fstop/resources"
	. "github.com/eburlingame/fstop/utils"

	"github.com/gin-gonic/gin"
)

func AlbumsListGetHandler(r *Resources) gin.HandlerFunc {
	type AlbumElement struct {
		AlbumId      string
		Slug         string
		Name         string
		Description  string
		CoverImageId string
		LatestDate   string
		PublicURL    string
	}

	return func(c *gin.Context) {
		albumListings, _ := r.Db.ListAlbumsCovers(true, 400, 100, 0)

		albums := []AlbumElement{}

		for i := range albumListings {
			albums = append(albums, AlbumElement{
				AlbumId:      albumListings[i].AlbumId,
				Slug:         albumListings[i].Slug,
				Name:         albumListings[i].Name,
				Description:  albumListings[i].Description,
				CoverImageId: albumListings[i].CoverImageId,
				LatestDate:   albumListings[i].LatestDate,
				PublicURL:    PublicImageURL(r.Config.S3BaseUrl, albumListings[i].File.StoragePath),
			})

		}

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

			if len(img.Files) == 0 {
				log.Printf("No files found for image %s\n", img.ImageId)
				continue
			}

			smallImage := FindSizedImage(img.Files, 500)

			imagesWithSrcSets = append(imagesWithSrcSets, ImageWithSrcSet{
				ImageId:       img.ImageId,
				SrcSet:        ComputeImageSrcSet(r.Config.S3BaseUrl, img.Files),
				SmallImageUrl: PublicImageURL(r.Config.S3BaseUrl, smallImage.StoragePath),
				Width:         img.WidthPixels,
				Height:        img.HeightPixels,
				Title:         img.DateTimeOriginal.Format("Monday, January _2, 2006"),
				Description:   GetAlbumImageCameraAndMetaDescription(&img),
			})
		}

		c.HTML(http.StatusOK, "album.html", gin.H{
			"album":  album,
			"images": imagesWithSrcSets,
		})
	}
}
