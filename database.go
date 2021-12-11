package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Table structs

type Image struct {
	gorm.Model
	Filename string
}

type Database interface {
	AddImage(image *Image) error
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

	base := &SqliteDatabase{
		db: db,
	}

	return base, nil
}

func (d *SqliteDatabase) AddImage(image *Image) error {
	// Create
	d.db.Create(image)

	return nil
}
