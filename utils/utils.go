package utils

import (
	"fmt"
	"strings"

	"path/filepath"

	. "github.com/eburlingame/fstop/models"

	"github.com/google/uuid"
)

func Uuid() string {
	uuidWithHyphen := uuid.New()
	return strings.Replace(uuidWithHyphen.String(), "-", "", -1)
}

func GetExtension(filename string) string {
	return filepath.Ext(filename)
}

func GetLongestEdge(width int, height int) int {
	if width > height {
		return width
	}
	return height
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

func ComputeImageSrcSet(files []File) string {
	srcs := []string{}

	for _, file := range files {
		srcs = append(srcs, fmt.Sprintf("%s %dw", file.PublicURL, file.Width))
	}

	return strings.Join(srcs, ", ")
}

func FindSizedImage(files []File, minWidth int) *File {
	fmt.Printf("files: %?", files)

	if len(files) == 0 {
		fmt.Println("Empty files")
		return nil
	}

	for _, file := range files {
		if file.Width > uint64(minWidth) {
			return &file
		}
	}

	largestFile := files[len(files)-1]
	return &largestFile
}
