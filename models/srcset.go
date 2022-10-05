package models

type ImageWithSrcSet struct {
	ImageId       string
	SrcSet        string
	SmallImageUrl string // The public URL where the file is available
	Width         uint64 // Width in pixels of the image file
	Height        uint64 // Height in pixels of the image file
	Title         string
	Description   string
}
