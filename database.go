package main

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Table structs

type Database interface {
	GetImage(image *Image, fileId string) error
	GetImagesInImportBatch(images *[]Image, batchId string)
	AddImage(image *Image) error
	UpdateImageProcessedStatus(fileId string, isProcessed bool) error

	AddImportBatch() string

	AddFile(file *File) error
	GetFile(file *File, fileId string, minWidth int) error
}

type SqliteDatabase struct {
	db *gorm.DB
}

func InitSqliteDatabase(config *Configuration) (*SqliteDatabase, error) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Image{})
	db.AutoMigrate(&File{})
	db.AutoMigrate(&ImportBatch{})

	base := &SqliteDatabase{
		db: db,
	}

	return base, nil
}

func (d *SqliteDatabase) GetImage(image *Image, fileId string) error {
	d.db.First(&image, "file_id = ?", fileId)
	return nil
}

func (d *SqliteDatabase) AddImage(image *Image) error {
	d.db.Create(image)

	return nil
}

func (d *SqliteDatabase) UpdateImageProcessedStatus(fileId string, isProcessed bool) error {
	var image Image

	d.db.First(&image, "file_id = ?", fileId)
	d.db.Model(&image).Update("IsProcessed", isProcessed)

	return nil
}

func (d *SqliteDatabase) AddFile(file *File) error {
	d.db.Create(file)

	return nil
}

func (d *SqliteDatabase) GetFile(file *File, fileId string, minWidth int) error {
	d.db.Order("width asc").Where("file_id = ? AND width > ? AND is_original = ?", fileId, minWidth, false).First(file)
	return nil
}

func (d *SqliteDatabase) AddImportBatch() string {
	id := Uuid()

	d.db.Create(&ImportBatch{
		Id:   id,
		Date: time.Now(),
	})

	return id
}

func (d *SqliteDatabase) GetImagesInImportBatch(images *[]Image, batchId string) {
	d.db.Where("import_batch_id = ?", batchId).Find(&images)
}
