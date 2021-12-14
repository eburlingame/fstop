package handlers

import (
	"fmt"
	"net/http"
	"strings"

	. "github.com/eburlingame/fstop/models"
	. "github.com/eburlingame/fstop/process"
	. "github.com/eburlingame/fstop/resources"
	. "github.com/eburlingame/fstop/utils"

	"github.com/gin-gonic/gin"
)

func AdminImportGet(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		files, err := r.Storage.ListFiles(r.Config.S3UploadFolder)
		if err != nil {
			c.String(500, "Error listing files: %s\n", err)
			return
		}

		for i := range files {
			files[i] = strings.Replace(files[i], r.Config.S3UploadFolder+"/", "", 1)
		}

		c.HTML(http.StatusOK, "import.html", gin.H{
			"files": files,
		})
	}
}

func AdminImportPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		names, _ := c.GetPostFormArray("names")

		importBatchId := Uuid()

		images := []ImageImport{}

		for _, value := range names {
			images = append(images, ImageImport{
				ImageId:        Uuid(),
				ImportBatchId:  importBatchId,
				UploadFilePath: r.Config.S3UploadFolder + "/" + value,

				Sizes: []OutputImageSize{
					{
						LongEdge:    200,
						Suffix:      "_thumb",
						Format:      "png",
						Extension:   ".png",
						ContentType: "image/png",
					},
					{
						LongEdge:    600,
						Suffix:      "_small",
						Format:      "png",
						Extension:   ".png",
						ContentType: "image/png",
					},
					{
						LongEdge:    1080,
						Suffix:      "_medium",
						Format:      "png",
						Extension:   ".png",
						ContentType: "image/png",
					},
					{
						LongEdge:    1920,
						Suffix:      "_large",
						Format:      "png",
						Extension:   ".png",
						ContentType: "image/png",
					},
				},
			})
		}

		batch := ImportBatchRequest{
			ImportBatchId: importBatchId,
			Images:        images,
		}

		go ImportImageBatch(r, batch)

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

		fmt.Printf("Batchid: %s\n", params.BatchId)

		var images []Image
		r.Db.GetImagesInImportBatch(&images, params.BatchId)
		fmt.Println(images)

		type Status struct {
			IsProcessed bool
			URL         string
		}

		statuses := make([]Status, len(images))
		allProcessed := true

		for i, img := range images {
			statuses[i].IsProcessed = img.IsProcessed

			if img.IsProcessed {
				var file File

				r.Db.GetFile(&file, img.ImageId, 100)
				statuses[i].URL = file.PublicURL
			} else {
				allProcessed = false
			}
		}

		c.HTML(http.StatusOK, "import_status_table.html", gin.H{
			"poll":          !allProcessed || len(statuses) == 0,
			"statuses":      statuses,
			"importBatchId": params.BatchId,
		})
	}
}
