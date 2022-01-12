package models

type ImageImportTask struct {
	ImageId       string `gorm:"primarykey"`
	ImportBatchId string `gorm:"primarykey"`
	Filename      string
	IsProcessed   bool
}
