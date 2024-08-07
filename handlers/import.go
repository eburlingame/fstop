package handlers

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	. "github.com/eburlingame/fstop/models"
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

func getImportSizes() []OutputImageSize {
	return []OutputImageSize{
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
	}
}

func performImport(r *Resources, names []string, albumId string) string {
	importBatchId := Uuid()
	images := []ImageImport{}

	for _, value := range names {
		images = append(images, ImageImport{
			InitialImport:   true,
			ImageId:         Uuid(),
			ImportBatchId:   importBatchId,
			AlbumId:         albumId,
			OriginalFileKey: r.Config.S3UploadFolder + "/" + value,
			Sizes:           getImportSizes(),
		})
	}

	for _, image := range images {
		r.Db.AddImageImport(image.ImportBatchId, image.ImageId, filepath.Base(image.OriginalFileKey))
	}

	for i := range images {
		r.Queue.AddTask(images[i])
	}

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

func SingleResizeApiPostHandler(r *Resources) gin.HandlerFunc {
	type ResizeRequest struct {
		ImageIds []string `json:"imageIds"`
	}

	return func(c *gin.Context) {
		var resizeRequest ResizeRequest

		err := c.Bind(&resizeRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unrecognized payload: %s", err)})
			return
		}

		log.Printf("Resizing %d images\n", len(resizeRequest.ImageIds))

		importBatchId := Uuid()

		allFiles := []File{}
		err = r.Db.ListOriginalImageFiles(&allFiles)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Error listing images: %s", err),
			})
			return
		}

		files := []File{}
		for _, image := range allFiles {
			for _, id := range resizeRequest.ImageIds {
				if image.ImageId == id && image.IsOriginal {
					files = append(files, image)
				}
			}
		}

		for _, file := range files {
			r.Queue.AddTask(ImageImport{
				InitialImport:   false,
				ImageId:         file.ImageId,
				ImportBatchId:   importBatchId,
				AlbumId:         "",
				OriginalFileKey: file.StoragePath,
				Sizes:           getImportSizes(),
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"batchId": importBatchId,
		})
	}
}

func BulkResizeApiPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		importBatchId := Uuid()
		imports := []ImageImport{}

		files := []File{}
		err := r.Db.ListOriginalImageFiles(&files)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Error listing images: %s", err),
			})
			return
		}

		for _, file := range files {
			imports = append(imports, ImageImport{
				InitialImport:   false,
				ImageId:         file.ImageId,
				ImportBatchId:   importBatchId,
				AlbumId:         "",
				OriginalFileKey: file.StoragePath,
				Sizes:           getImportSizes(),
			})
		}

		for i := range imports {
			r.Queue.AddTask(imports[i])
		}

		c.JSON(http.StatusOK, gin.H{
			"batchId": importBatchId,
		})
	}
}

// PurgeOrphanImagesApiPostHandler deletes images in S3 that are not referenced in the database
func PurgeOrphanImagesApiPostHandler(r *Resources) gin.HandlerFunc {
	return func(c *gin.Context) {
		all_files := []File{}
		err := r.Db.ListFiles(&all_files)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Error listing images: %s", err),
			})
			return
		}

		fileKeys := map[string]bool{}
		for _, file := range all_files {
			fileKeys[file.StoragePath] = true
		}

		storage_files, err := r.Storage.ListFiles(r.Config.S3MediaFolder)
		if err != nil {
			log.Printf("Error listing files: %s\n", err)
			return
		}

		orphan_keys := []string{}
		for _, stored_file := range storage_files {
			if !fileKeys[stored_file] {
				log.Printf("Purging orphaned file: %s\n", stored_file)
				orphan_keys = append(orphan_keys, stored_file)

				err := r.Storage.DeleteFile(stored_file)
				if err != nil {
					log.Printf("Error deleting file: %s\n", err)
				}
			}
		}

		log.Printf("Purged %d orphaned files\n", len(orphan_keys))

		c.JSON(http.StatusOK, gin.H{
			"purges": orphan_keys,
		})
	}
}
