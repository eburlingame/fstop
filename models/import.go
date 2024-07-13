package models

type ImageImportTask struct {
	ImageId       string `gorm:"primarykey"`
	ImportBatchId string `gorm:"primarykey"`
	Filename      string
	IsProcessed   bool
}

type OutputImageSize struct {
	LongEdge    int
	Quality     int
	Suffix      string
	Extension   string
	Format      string
	ContentType string
}

type ImageImport struct {
	InitialImport  bool
	ImageId        string
	ImportBatchId  string
	UploadFilePath string
	AlbumId        string
	Sizes          []OutputImageSize
}
