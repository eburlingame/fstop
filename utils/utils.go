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

func PublicImageURL(s3Url string, storagePath string) string {
	return fmt.Sprintf("%s/%s", s3Url, storagePath)
}

func ComputeImageSrcSet(s3URL string, files []File) string {
	srcs := []string{}

	for _, file := range files {
		if strings.HasSuffix(file.StoragePath, ".webp") && !file.IsOriginal {
			srcs = append(srcs, fmt.Sprintf("%s %dw", PublicImageURL(s3URL, file.StoragePath), file.Width))
		}
	}

	return strings.Join(srcs, ", ")
}

func FindSizedImage(files []File, minWidth int) *File {
	if len(files) == 0 {
		return nil
	}

	for _, file := range files {
		if file.Width > uint64(minWidth) && strings.HasSuffix(file.StoragePath, ".webp") {
			return &file
		}
	}

	largestFile := files[len(files)-1]
	return &largestFile
}
