package models

import "time"

type Album struct {
	AlbumId      string `gorm:"primarykey"`
	Slug         string
	Name         string
	Description  string
	CoverImageId string
	IsPublished  bool
}

type AlbumImage struct {
	AlbumId string `gorm:"primarykey"`
	ImageId string `gorm:"primarykey"`
}

// Matches the AlbumWithImage view
type AlbumWithImage struct {
	ImageId          string
	WidthPixels      uint64
	HeightPixels     uint64
	DateTimeOriginal time.Time
	CameraModel      string
	Lens             string
	ShutterSpeed     string
	FNumber          float64
	ISO              float64
	FocalLength      string
	Files            []File `gorm:"foreignKey:ImageId;references:ImageId"`
}
