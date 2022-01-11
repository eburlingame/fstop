package resources

import (
	"database/sql"
	"log"
	"os"
	"time"

	. "github.com/eburlingame/fstop/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Table structs

type Database interface {
	GetImage(image *Image, imageId string) error
	GetImagesInImportBatch(images *[]Image, batchId string)
	AddImage(image *Image) error
	DeleteImage(imageId string) error
	UpdateImageProcessedStatus(imageId string, isProcessed bool) error

	ListLatestPhotos(minWidth int, limit int, offset int) ([]File, error)

	AddFile(file *File) error
	GetFile(file *File, fileId string, minWidth int) error
	ListImageFiles(file *[]File, imageId string) error

	ListAlbums(album *[]Album) error
	ListAlbumsCovers(albums *[]AlbumFile, publishedOnly bool, minWidth int, limit int, offset int) error

	GetAlbum(album *Album, albumId string) error
	GetAlbumBySlug(album *Album, albumSlug string) error
	AddAlbum(album Album) error
	DeleteAlbum(albumId string) error
	UpdateAlbum(albumId string, updatedAlbum *Album) error
	AddImageToAlbum(albumId string, imageId string) error
	RemoveImageFromAlbum(albumId string, imageId string) error
	ListAlbumImages(files *[]File, albumSlug string, minWidth int, limit int, offset int) error
}

type SqliteDatabase struct {
	db *gorm.DB
}

const ImageFileView string = `
	DROP VIEW IF EXISTS image_with_files;

	CREATE VIEW image_with_files AS
		SELECT 
			*,
			(SELECT MIN(f2.width) FROM files f2 WHERE f2.image_id = image_id) AS min_width,
			(SELECT MAX(f2.width) FROM files f2 WHERE f2.image_id = image_id) AS max_width
		FROM images
		JOIN (
			SELECT * 
			FROM files f
		) as files ON files.image_id  = images.image_id 
		WHERE images.is_processed = TRUE
		ORDER BY date_time_original DESC;
`

const AlbumsAndImagesView string = `
	DROP VIEW IF EXISTS albums_with_images;

	CREATE VIEW albums_with_images AS
		SELECT *
		FROM album_images ai
		JOIN albums a ON ai.album_id  = a.album_id
		JOIN images i ON i.image_id  = ai.image_id;
`

const AlbumComputed string = `
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
				ELSE (SELECT aai.image_id 
						FROM albums_with_images aai 
						WHERE aai.album_id = a.album_id 
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

const AlbumImageFiles string = `
	DROP VIEW IF EXISTS album_images_with_files;

	CREATE VIEW album_images_with_files AS
		SELECT *
		FROM image_with_files if2
		LEFT JOIN albums_with_images aai ON aai.image_id = if2.image_id
		WHERE aai.album_id IS NOT NULL
`

func InitSqliteDatabase(config *Configuration) (*SqliteDatabase, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
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

	db.Exec(ImageFileView)
	db.Exec(AlbumsAndImagesView)
	db.Exec(AlbumComputed)
	db.Exec(AlbumImageFiles)

	base := &SqliteDatabase{
		db: db,
	}

	return base, nil
}

func (d *SqliteDatabase) GetImage(image *Image, fileId string) error {
	d.db.First(&image, "image_id = ?", fileId)
	return nil
}

func (d *SqliteDatabase) AddImage(image *Image) error {
	d.db.Create(image)

	return nil
}

func (d *SqliteDatabase) DeleteImage(imageId string) error {
	d.db.Where("image_id = ?", imageId).Delete(&Image{})
	d.db.Where("image_id = ?", imageId).Delete(&AlbumImage{})
	d.db.Where("image_id = ?", imageId).Delete(&File{})

	return nil
}

func (d *SqliteDatabase) UpdateImageProcessedStatus(fileId string, isProcessed bool) error {
	var image Image

	d.db.First(&image, "image_id = ?", fileId)
	d.db.Model(&image).Update("IsProcessed", isProcessed)

	return nil
}

func (d *SqliteDatabase) AddFile(file *File) error {
	d.db.Create(file)

	return nil
}

func (d *SqliteDatabase) ListLatestPhotos(minWidth int, limit int, offset int) ([]File, error) {
	var images []Image

	d.db.Preload("Files").
		Limit(limit).
		Offset(offset).
		Find(&images).
		Order("date_time_original ASC, width ASC")

	sizedFiles := []File{}

	for _, image := range images {
		sizeFound := false

		for _, file := range image.Files {
			if file.Width > uint64(minWidth) {
				sizedFiles = append(sizedFiles, file)
				sizeFound = true
				break
			}
		}

		if !sizeFound {
			largestFile := image.Files[len(image.Files)-1]
			sizedFiles = append(sizedFiles, largestFile)
		}
	}

	return sizedFiles, nil
}

type AlbumFile struct {
	AlbumId      string
	Slug         string
	Name         string
	Description  string
	CoverImageId string
	PublicURL    string
}

func (d *SqliteDatabase) ListAlbumsCovers(albums *[]AlbumFile, publishedOnly bool, minWidth int, limit int, offset int) error {
	published := 0
	if publishedOnly {
		published = 1
	}

	d.db.Raw(`
		SELECT 
			album_id,
			slug, 
			name, 
			description,
			latest_date,
			cover_image_id,
			public_url
		FROM album_covers a
		INNER JOIN (SELECT i.image_id, public_url
					FROM files f
					JOIN images i ON i.image_id = f.image_id
					WHERE width = (
								SELECT MIN(width) 
								FROM image_with_files if1
								WHERE if1.image_id = f.image_id 
								AND if1.width > @minWidth OR if1.max_width < @minWidth 
								)
					) AS small_images
		ON small_images.image_id = cover_image_id
		WHERE a.is_published >= @published
		ORDER BY latest_date DESC
		LIMIT @limit
		OFFSET @offset

		`,
		sql.Named("minWidth", minWidth),
		sql.Named("limit", limit),
		sql.Named("offset", offset),
		sql.Named("published", published)).
		Scan(&albums)

	return nil
}

func (d *SqliteDatabase) GetFile(file *File, fileId string, minWidth int) error {
	d.db.
		Order("width asc").
		Where("image_id = ? AND width > ?", fileId, minWidth).
		First(file)

	return nil
}

func (d *SqliteDatabase) ListImageFiles(files *[]File, imageId string) error {
	d.db.
		Order("width asc").
		Where("image_id = ?", imageId).
		Find(files)

	return nil
}

func (d *SqliteDatabase) GetImagesInImportBatch(images *[]Image, batchId string) {
	d.db.Where("import_batch_id = ?", batchId).Find(&images)
}

func (d *SqliteDatabase) AddAlbum(album Album) error {
	d.db.Create(&album)

	return nil
}

func (d *SqliteDatabase) DeleteAlbum(albumId string) error {
	d.db.Where("album_id = ?", albumId).Delete(&Album{})
	d.db.Where("album_id = ?", albumId).Delete(&AlbumImage{})

	return nil
}

func (d *SqliteDatabase) UpdateAlbum(albumId string, updatedAlbum *Album) error {
	d.db.Model(&Album{}).
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
	d.db.Find(&album, "album_id = ?", albumId)
	return nil
}

func (d *SqliteDatabase) GetAlbumBySlug(album *Album, albumSlug string) error {
	d.db.Find(&album, "slug = ?", albumSlug)
	return nil
}

func (d *SqliteDatabase) ListAlbums(album *[]Album) error {
	d.db.Find(&album)
	return nil
}

func (d *SqliteDatabase) AddImageToAlbum(albumId string, imageId string) error {
	d.db.Create(&AlbumImage{
		AlbumId: albumId,
		ImageId: imageId,
	})
	return nil
}

func (d *SqliteDatabase) RemoveImageFromAlbum(albumId string, imageId string) error {
	d.db.Where("album_id = ? AND image_id = ?", albumId, imageId).Delete(&AlbumImage{})
	return nil
}

func (d *SqliteDatabase) ListAlbumImages(files *[]File, albumSlug string, minWidth int, limit int, offset int) error {
	d.db.Raw(`
		SELECT * 
		FROM album_images_with_files aif 
		WHERE  
			width = (SELECT MIN(width) FROM image_with_files if2 WHERE if2.image_id = aif.image_id AND if2.width > 400) AND
			slug = ?
		ORDER BY date_time_original DESC;
	`, albumSlug).Find(files)

	return nil
}
