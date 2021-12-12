package main

import (
	"strings"

	"path/filepath"

	"github.com/google/uuid"
)

func Uuid() string {
	uuidWithHyphen := uuid.New()
	return strings.Replace(uuidWithHyphen.String(), "-", "", -1)
}

func GetExtension(filename string) string {
	return filepath.Ext(filename)
}

func ResizeLongEdgeDimensions(width int, height int, longEdge int) (int, int) {
	aspectRatio := float32(width) / float32(height)

	if width >= height {
		if longEdge > width {
			return width, height
		}

		return longEdge, int(float32(longEdge) / aspectRatio)
	} else {
		if longEdge > height {
			return width, height
		}

		return int(float32(longEdge) * aspectRatio), longEdge
	}
}
