package handlers

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	. "github.com/eburlingame/fstop/models"
	. "github.com/eburlingame/fstop/process"
	. "github.com/eburlingame/fstop/resources"
	. "github.com/eburlingame/fstop/utils"
	"github.com/gosimple/slug"

	"github.com/gin-gonic/gin"
)

func ImportSelectionPage(r *Resources, c *gin.Context, formError error) {
	files, err := r.Storage.ListFiles(r.Config.S3UploadFolder)
	if err != nil {
		c.String(500, "Error listing files: %s\n", err)
		return
	}

	var albums []Album
	r.Db.ListAlbums(&albums)

	for i := range files {
		files[i] = strings.Replace(files[i], r.Config.S3UploadFolder+"/", "", 1)
	}

	c.HTML(http.StatusOK, "import.html", gin.H{
		"files":     files,
		"albums":    albums,
		"hasAlbums": len(albums) > 0,
		"hasError":  formError != nil,
		"error":     fmt.Sprintf("Error: %s", formError),
	})
}

func AdminImportGet(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		ImportSelectionPage(r, c, nil)
	}
}

func getFormAlbumId(r *Resources, c *gin.Context) (string, error) {
	addToAlbum, _ := c.GetPostForm("addToAlbum")
	albumSelection, _ := c.GetPostForm("albumSelection")
	newAlbumName, _ := c.GetPostForm("newAlbumName")
	existingAlbumId, _ := c.GetPostForm("existingAlbumId")

	var album Album
	albumId := ""

	if addToAlbum == "on" {
		if albumSelection == "existing" {
			r.Db.GetAlbum(&album, existingAlbumId)
			if album.AlbumId == "" {
				return "", fmt.Errorf("Unexpected type %s", albumSelection)
			}

			albumId = album.AlbumId
		} else if albumSelection == "new" {
			if strings.Trim(newAlbumName, " ") == "" {
				log.Printf("Name cannot be empty")
				return "", fmt.Errorf("Name cannot be empty")
			}

			albumId = Uuid()
			r.Db.AddAlbum(Album{
				AlbumId:      albumId,
				Name:         newAlbumName,
				Slug:         slug.Make(newAlbumName),
				Description:  "",
				CoverImageId: "",
				IsPublished:  true,
			})
		} else {
			return "", fmt.Errorf("Unexpected type %s", albumSelection)
		}
	}

	return albumId, nil
}

func performImport(r *Resources, names []string, albumId string) string {
	importBatchId := Uuid()
	images := []ImageImport{}

	for _, value := range names {
		images = append(images, ImageImport{
			ImageId:        Uuid(),
			ImportBatchId:  importBatchId,
			AlbumId:        albumId,
			UploadFilePath: r.Config.S3UploadFolder + "/" + value,

			Sizes: []OutputImageSize{
				{
					LongEdge:    200,
					Quality:     80,
					Suffix:      "_thumb",
					Format:      "webp",
					Extension:   ".webp",
					ContentType: "image/webp",
				},
				{
					LongEdge:    600,
					Quality:     80,
					Suffix:      "_small",
					Format:      "webp",
					Extension:   ".webp",
					ContentType: "image/webp",
				},
				{
					LongEdge:    1080,
					Quality:     80,
					Suffix:      "_medium",
					Format:      "webp",
					Extension:   ".webp",
					ContentType: "image/webp",
				},
				{
					LongEdge:    1920,
					Quality:     65,
					Suffix:      "_large",
					Format:      "webp",
					Extension:   ".webp",
					ContentType: "image/webp",
				},
				{
					LongEdge:    2560,
					Quality:     50,
					Suffix:      "_xlarge",
					Format:      "webp",
					Extension:   ".webp",
					ContentType: "image/webp",
				},
				// {
				// 	LongEdge:    10000,
				// 	Quality:     50,
				// 	Suffix:      "_original",
				// 	Format:      "webp",
				// 	Extension:   ".webp",
				// 	ContentType: "image/webp",
				// },
			},
		})
	}

	for _, image := range images {
		r.Db.AddImageImport(image.ImportBatchId, image.ImageId, filepath.Base(image.UploadFilePath))
	}

	batch := ImportBatchRequest{
		ImportBatchId: importBatchId,
		Images:        images,
	}

	go ImportImageBatch(r, batch)

	return importBatchId
}

type ImportStatus struct {
	IsProcessed bool   `json:"isProcessed"`
	Filename    string `json:"filename"`
	URL         string `json:"url"`
}

func getImportStatuses(r *Resources, importBatchId string) (bool, []ImportStatus) {
	var images []ImageImportTask
	r.Db.GetImagesInImportBatch(&images, importBatchId)

	statuses := make([]ImportStatus, len(images))
	allProcessed := true

	for i, img := range images {
		statuses[i].IsProcessed = img.IsProcessed
		statuses[i].Filename = img.Filename

		if img.IsProcessed {
			var file File
			r.Db.GetFile(&file, img.ImageId, 100)

			statuses[i].URL = PublicImageURL(r.Config.S3BaseUrl, file.StoragePath)
		} else {
			allProcessed = false
		}
	}

	return allProcessed, statuses
}

func AdminImportPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		names, _ := c.GetPostFormArray("names")

		albumId, err := getFormAlbumId(r, c)
		if err != nil {
			ImportSelectionPage(r, c, err)
			return
		}

		importBatchId := performImport(r, names, albumId)

		c.HTML(200, "import_complete.html", gin.H{
			"importBatchId": importBatchId,
		})
	}
}

func AdminImportStatusGetHandler(r *Resources) gin.HandlerFunc {
	type UriParams struct {
		BatchId string `uri:"batchId" binding:"required"`
	}

	return func(c *gin.Context) {
		var params UriParams

		err := c.BindUri(&params)
		if err != nil {
			c.Status(404)
			return
		}

		allProcessed, statuses := getImportStatuses(r, params.BatchId)

		c.HTML(http.StatusOK, "import_status_table.html", gin.H{
			"poll":          !allProcessed || len(statuses) == 0,
			"statuses":      statuses,
			"importBatchId": params.BatchId,
		})
	}
}

type ImportApiRequest struct {
	Names           []string `json:"names"`
	NewAlbumName    string   `json:"newAlbumName,omitempty"`
	ExistingAlbumId string   `json:"existingAlbumId,omitempty"`
}

func ImportApiPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		var importRequest ImportApiRequest

		err := c.Bind(&importRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unrecognized payload: %s", err)})
			return
		}

		albumId := ""

		if strings.Trim(importRequest.NewAlbumName, " ") != "" {
			albumId = Uuid()

			r.Db.AddAlbum(Album{
				AlbumId:      albumId,
				Name:         importRequest.NewAlbumName,
				Slug:         slug.Make(importRequest.NewAlbumName),
				Description:  "",
				CoverImageId: "",
				IsPublished:  true,
			})
		}

		if importRequest.ExistingAlbumId != "" {
			var album Album
			r.Db.GetAlbum(&album, importRequest.ExistingAlbumId)

			if album.AlbumId == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": fmt.Sprintf("Unknown album: %s", album.AlbumId),
				})
				return
			}

			albumId = importRequest.ExistingAlbumId
		}

		batchId := performImport(r, importRequest.Names, albumId)

		c.JSON(200, gin.H{
			"importBatchId": batchId,
		})
	}
}

func ImportStateApiGetHandler(r *Resources) gin.HandlerFunc {
	type UriParams struct {
		BatchId string `uri:"batchId" binding:"required"`
	}

	return func(c *gin.Context) {
		var params UriParams

		err := c.BindUri(&params)
		if err != nil {
			c.Status(404)
			return
		}

		allProcessed, statuses := getImportStatuses(r, params.BatchId)

		c.JSON(http.StatusOK, gin.H{
			"batchId":      params.BatchId,
			"allProcessed": allProcessed,
			"statuses":     statuses,
		})
	}
}
