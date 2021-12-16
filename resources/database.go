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
	UpdateImageProcessedStatus(imageId string, isProcessed bool) error

	ListLatestPhotos(photos *[]File, minWidth int, limit int, offset int) error

	AddFile(file *File) error
	GetFile(file *File, fileId string, minWidth int) error

	ListAlbums(album *[]Album) error
	ListAlbumsCovers(albums *[]Alb, minWidth int, limit int, offset int) error
	GetAlbum(album *Album, albumId string) error
	GetAlbumByName(album *Album, albumName string) error
	AddAlbum(album Album) error
	AddImageToAlbum(albumId string, imageId string) error
}

type SqliteDatabase struct {
	db *gorm.DB
}

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

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
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

func (d *SqliteDatabase) ListLatestPhotos(files *[]File, minWidth int, limit int, offset int) error {
	d.db.Raw(`
		SELECT *
		FROM files
		JOIN images i ON i.image_id  = files.image_id
		WHERE width = (SELECT MIN(width) 
			FROM files f 
			WHERE f.image_id = files.image_id 
			AND f.width > ? OR 
				(SELECT MAX(width) FROM files f1 WHERE f1.image_id = files.image_id) < ?
			)
		ORDER BY i.date_time_original DESC
		LIMIT ?
		OFFSET ?
		`, minWidth, minWidth, limit, offset).
		Find(files)

	return nil
}

type Alb struct {
	Id           string `gorm:"column:id"`
	Slug         string `gorm:"column:slug"`
	Name         string `gorm:"column:name"`
	Description  string `gorm:"column:description"`
	CoverImageId string `gorm:"column:cover_image_id"`
	PublicURL    string `gorm:"column:public_url"`
}

func (d *SqliteDatabase) ListAlbumsCovers(albums *[]Alb, minWidth int, limit int, offset int) error {
	d.db.Raw(`
		SELECT 
			id,
			slug, 
			name, 
			description,
			latest_date,
			cover_image_id,
			public_url
		FROM 
			(SELECT 
				a.id,
				a.slug, 
				a.description,
				a.name,
				(CASE WHEN a.cover_image_id <> "" 
					THEN a.cover_image_id 
					ELSE (SELECT i1.image_id 
							FROM album_images ai 
							INNER JOIN images i1 
							ON i1.image_id = ai.image_id AND ai.album_id = a.id 
							LIMIT 1)
				END) AS cover_image_id,
				(SELECT 
					MAX(date_time_original) 
					FROM album_images ai2
					INNER JOIN images i2 
					ON i2.image_id = ai2.image_id 
					WHERE ai2.album_id = a.id) AS latest_date
			FROM albums a) AS a
		INNER JOIN (SELECT i.image_id, public_url
					FROM files f
					JOIN images i ON i.image_id = f.image_id
					WHERE width = (
								SELECT MIN(width) 
								FROM files f1
								WHERE f1.image_id = f.image_id 
								AND f1.width > @minWidth  OR (SELECT MAX(width) FROM files f2 WHERE f1.image_id = i.image_id) < @minWidth 
								)
					) AS small_images 
		ON small_images.image_id = cover_image_id
		ORDER BY latest_date DESC
		LIMIT @limit
		OFFSET @offset

		`, sql.Named("minWidth", minWidth), sql.Named("limit", limit), sql.Named("offset", offset)).
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

func (d *SqliteDatabase) GetImagesInImportBatch(images *[]Image, batchId string) {
	d.db.Where("import_batch_id = ?", batchId).Find(&images)
}

func (d *SqliteDatabase) AddAlbum(album Album) error {
	d.db.Create(&album)

	return nil
}

func (d *SqliteDatabase) DeleteAlbum(albumId string) error {
	d.db.Delete(&Album{}, albumId)
	d.db.Where("album_id = ?", albumId).Delete(&AlbumImage{})

	return nil
}

func (d *SqliteDatabase) UpdateAlbum(albumId string, updatedAlbum *Album) error {
	d.db.Where("album_id = ?", albumId).Updates(updatedAlbum)

	return nil
}

func (d *SqliteDatabase) GetAlbum(album *Album, albumId string) error {
	d.db.Find(&album, "id = ?", albumId)
	return nil
}

func (d *SqliteDatabase) GetAlbumByName(album *Album, albumName string) error {
	d.db.Find(&album, "name = ?", albumName)
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
