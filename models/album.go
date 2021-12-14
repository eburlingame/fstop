package models

type Album struct {
	Id          string `gorm:"primarykey"`
	Name        string // The uuid for the image
	Description string // The id of the batch where the file was uploaded
}

type AlbumImage struct {
	Id      string `gorm:"primarykey"`
	ImageId string // The uuid for the image
	Order   int
}
