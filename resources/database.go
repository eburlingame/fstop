package resources

import (
	"log"
	"os"
	"time"

	. "github.com/eburlingame/fstop/models"
	. "github.com/eburlingame/fstop/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Table structs

type Database interface {
	GetImage(image *Image, imageId string) error
	GetImagesInImportBatch(images *[]ImageImportTask, batchId string)
	AddImage(image *Image) error
	DeleteImage(imageId string) error

	AddImageImport(importBatchId string, imageId string, filename string) error
	UpdateImageProcessedStatus(imageId string, isProcessed bool) error

	ListLatestFiles(minWidth int, limit int, offset int) ([]File, error)
	ListLatestImages(limit int, offset int) ([]Image, error)

	AddFile(file *File) error
	GetFile(file *File, fileId string, minWidth int) error
	ListImageFiles(file *[]File, imageId string) error

	ListAlbums(album *[]Album) error
	ListAlbumsCovers(publishedOnly bool, minWidth int, limit int, offset int) ([]AlbumListing, error)

	GetAlbum(album *Album, albumId string) error
	GetAlbumBySlug(album *Album, albumSlug string) error
	AddAlbum(album Album) error
	DeleteAlbum(albumId string) error
	UpdateAlbum(albumId string, updatedAlbum *Album) error
	AddImageToAlbum(albumId string, imageId string) error
	RemoveImageFromAlbum(albumId string, imageId string) error
	ListAlbumImages(albumSlug string, minWidth int, limit int, offset int) ([]File, error)
	ListAlbumFiles(albumSlug string) ([]AlbumWithImage, error)
}

type SqliteDatabase struct {
	Db *gorm.DB
}

const AlbumWithImagesView string = `
	DROP VIEW IF EXISTS album_with_images;

	CREATE VIEW album_with_images AS
		SELECT *
		FROM album_images ai
		JOIN albums a ON ai.album_id = a.album_id
		JOIN images i ON i.image_id = ai.image_id;
`

const AlbumCovers string = `
	DROP VIEW IF EXISTS album_covers;

	CREATE VIEW album_covers AS
		SELECT 
			a.album_id,
			a.slug, 
			a.description,
			a.name,
			a.is_published,
			(CASE WHEN a.cover_image_id <> "" 
				THEN a.cover_image_id 
				ELSE (SELECT ai.image_id 
						FROM album_with_images ai 
						WHERE ai.album_id = a.album_id 
						ORDER BY date_time_original DESC
						LIMIT 1)
			END) AS cover_image_id,
			(SELECT 
				MAX(date_time_original) 
				FROM album_images ai2
				INNER JOIN images i2 
				ON i2.image_id = ai2.image_id 
				WHERE ai2.album_id = a.album_id) AS latest_date
		FROM albums a;
`

type AlbumCover struct {
	AlbumId      string
	Slug         string
	Name         string
	Description  string
	CoverImageId string
	PublicURL    string
	LatestDate   string

	Files []File `gorm:"foreignKey:ImageId;references:CoverImageId"`
}

func InitSqliteDatabase(config *Configuration) (*SqliteDatabase, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Warn, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,       // Disable color
		},
	)

	db, err := gorm.Open(sqlite.Open(config.SQLiteFilepath), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Image{})
	db.AutoMigrate(&File{})
	db.AutoMigrate(&Album{})
	db.AutoMigrate(&AlbumImage{})
	db.AutoMigrate(&ImageImportTask{})

	db.Exec(AlbumWithImagesView)
	db.Exec(AlbumCovers)

	base := &SqliteDatabase{
		Db: db,
	}

	return base, nil
}

func (d *SqliteDatabase) GetImage(image *Image, fileId string) error {
	d.Db.First(&image, "image_id = ?", fileId)
	return nil
}

func (d *SqliteDatabase) AddImage(image *Image) error {
	d.Db.Create(image)

	return nil
}

func (d *SqliteDatabase) DeleteImage(imageId string) error {
	d.Db.Where("image_id = ?", imageId).Delete(&Image{})
	d.Db.Where("image_id = ?", imageId).Delete(&AlbumImage{})
	d.Db.Where("image_id = ?", imageId).Delete(&File{})
	d.Db.Where("image_id = ?", imageId).Delete(&ImageImportTask{})

	return nil
}

func (d *SqliteDatabase) AddImageImport(importBatchId string, imageId string, filename string) error {
	d.Db.Create(&ImageImportTask{
		ImageId:       imageId,
		Filename:      filename,
		ImportBatchId: importBatchId,
		IsProcessed:   false,
	})

	return nil
}

func (d *SqliteDatabase) UpdateImageProcessedStatus(imageId string, isProcessed bool) error {
	var imageImport ImageImportTask

	d.Db.First(&imageImport, "image_id = ?", imageId)
	d.Db.Model(&imageImport).Update("IsProcessed", isProcessed)

	return nil
}

func (d *SqliteDatabase) AddFile(file *File) error {
	d.Db.Create(file)

	return nil
}

func preloadFilesQuery(db *gorm.DB) *gorm.DB {
	return db.Order("files.width ASC").Where("files.is_original = false")
}

func (d *SqliteDatabase) ListLatestFiles(minWidth int, limit int, offset int) ([]File, error) {
	var images []Image

	d.Db.Preload("Files", preloadFilesQuery).
		Limit(limit).
		Offset(offset).
		Order("date_time_original desc").
		Find(&images)

	sizedFiles := []File{}

	for _, image := range images {
		sizedImage := FindSizedImage(image.Files, minWidth)

		if sizedImage != nil {
			sizedFiles = append(sizedFiles, *sizedImage)
		}
	}

	return sizedFiles, nil
}

func (d *SqliteDatabase) ListLatestImages(limit int, offset int) ([]Image, error) {
	var images []Image

	d.Db.Preload("Files", preloadFilesQuery).
		Limit(limit).
		Offset(offset).
		Order("date_time_original desc").
		Find(&images)

	return images, nil
}

type AlbumListing struct {
	AlbumId      string
	Slug         string
	Name         string
	Description  string
	CoverImageId string
	LatestDate   string
	File         File
}

func (d *SqliteDatabase) ListAlbumsCovers(publishedOnly bool, minWidth int, limit int, offset int) ([]AlbumListing, error) {
	published := 0
	if publishedOnly {
		published = 1
	}

	var covers []AlbumCover

	d.Db.Preload("Files", preloadFilesQuery).
		Where("is_published >= ?", published).
		Order("latest_date DESC").
		Limit(limit).
		Offset(offset).
		Find(&covers)

	listings := []AlbumListing{}

	for _, cover := range covers {
		sizedImage := FindSizedImage(cover.Files, minWidth)

		if sizedImage != nil {
			listings = append(listings, AlbumListing{
				AlbumId:      cover.AlbumId,
				Slug:         cover.Slug,
				Name:         cover.Name,
				Description:  cover.Description,
				CoverImageId: cover.CoverImageId,
				LatestDate:   cover.LatestDate,
				File:         *sizedImage,
			})
		}
	}

	return listings, nil
}

func (d *SqliteDatabase) GetFile(file *File, fileId string, minWidth int) error {
	d.Db.
		Order("width asc").
		Where("image_id = ? AND width > ?", fileId, minWidth).
		First(file)

	return nil
}

func (d *SqliteDatabase) ListImageFiles(files *[]File, imageId string) error {
	d.Db.
		Order("width asc").
		Where("image_id = ?", imageId).
		Find(files)

	return nil
}

func (d *SqliteDatabase) GetImagesInImportBatch(images *[]ImageImportTask, batchId string) {
	d.Db.Where("import_batch_id = ?", batchId).Find(&images)
}

func (d *SqliteDatabase) AddAlbum(album Album) error {
	d.Db.Create(&album)

	return nil
}

func (d *SqliteDatabase) DeleteAlbum(albumId string) error {
	d.Db.Where("album_id = ?", albumId).Delete(&Album{})
	d.Db.Where("album_id = ?", albumId).Delete(&AlbumImage{})

	return nil
}

func (d *SqliteDatabase) UpdateAlbum(albumId string, updatedAlbum *Album) error {
	d.Db.Model(&Album{}).
		Where("album_id = ?", albumId).
		Updates(map[string]interface{}{
			"slug":           updatedAlbum.Slug,
			"name":           updatedAlbum.Name,
			"description":    updatedAlbum.Description,
			"cover_image_id": updatedAlbum.CoverImageId,
			"is_published":   updatedAlbum.IsPublished,
		})

	return nil
}

func (d *SqliteDatabase) GetAlbum(album *Album, albumId string) error {
	d.Db.Find(&album, "album_id = ?", albumId)
	return nil
}

func (d *SqliteDatabase) GetAlbumBySlug(album *Album, albumSlug string) error {
	d.Db.Find(&album, "slug = ?", albumSlug)
	return nil
}

func (d *SqliteDatabase) ListAlbums(album *[]Album) error {
	d.Db.Find(&album)
	return nil
}

func (d *SqliteDatabase) AddImageToAlbum(albumId string, imageId string) error {
	d.Db.Create(&AlbumImage{
		AlbumId: albumId,
		ImageId: imageId,
	})
	return nil
}

func (d *SqliteDatabase) RemoveImageFromAlbum(albumId string, imageId string) error {
	d.Db.Where("album_id = ? AND image_id = ?", albumId, imageId).Delete(&AlbumImage{})
	return nil
}

func (d *SqliteDatabase) ListAlbumImages(albumSlug string, minWidth int, limit int, offset int) ([]File, error) {
	var images []AlbumWithImage

	d.Db.Preload("Files", preloadFilesQuery).
		Where("slug = ?", albumSlug).
		Limit(limit).
		Offset(offset).
		Order("date_time_original DESC").
		Find(&images)

	sizedFiles := []File{}

	for _, image := range images {
		sizedImage := FindSizedImage(image.Files, minWidth)
		if sizedImage != nil {
			sizedFiles = append(sizedFiles, *sizedImage)
		}
	}

	return sizedFiles, nil
}

func (d *SqliteDatabase) ListAlbumFiles(albumSlug string) ([]AlbumWithImage, error) {
	var images []AlbumWithImage

	d.Db.Preload("Files", preloadFilesQuery).
		Where("slug = ?", albumSlug).
		Order("date_time_original DESC").
		Find(&images)

	return images, nil
}
