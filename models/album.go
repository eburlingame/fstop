package models

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
