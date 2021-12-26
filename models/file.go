package models

type File struct {
	FileId        string `gorm:"primarykey"`
	ImageId       string // The uuid for the image
	ImportBatchId string // The id of the batch where the file was uploaded
	Filename      string // The filename with extension
	StoragePath   string // The path to the file, withing the storage bucket
	PublicURL     string // The public URL where the file is available
	IsOriginal    bool   // True if this is an original file
	Width         uint64 // Width in pixels of the image file
	Height        uint64 // Height in pixels of the image file
}
