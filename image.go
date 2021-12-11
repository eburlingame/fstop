package main

import (
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

const exifTag = "exifTag"

type Image struct {
	gorm.Model
	Filename string

	// EXIF data
	ShutterSpeed                   string `exifTag:"Shutter Speed"`
	Manufacturer                   string `exifTag:"Manufacturer"`
	CfaPattern                     string `exifTag:"CFA Pattern"`
	MeteringMode                   string `exifTag:"Metering Mode"`
	OffsetTimeForDateTimeOriginal  string `exifTag:"Offset Time For DateTimeOriginal"`
	FocalPlaneXResolution          string `exifTag:"Focal Plane X-Resolution"`
	ExposureMode                   string `exifTag:"Exposure Mode"`
	RecommendedExposureIndex       string `exifTag:"Recommended Exposure Index"`
	FNumber                        string `exifTag:"F-Number"`
	SceneCaptureType               string `exifTag:"Scene Capture Type"`
	OffsetTimeForDateTimeDigitized string `exifTag:"Offset Time For DateTimeDigitized"`
	OffsetTimeForDateTime          string `exifTag:"Offset Time For DateTime"`
	ResolutionUnit                 string `exifTag:"Resolution Unit"`
	WhiteBalance                   string `exifTag:"White Balance"`
	FileSource                     string `exifTag:"File Source"`
	SubSecondTimeDigitized         string `exifTag:"Sub-second Time (Digitized)"`
	ExposureProgram                string `exifTag:"Exposure Program"`
	BodySerialNumber               string `exifTag:"Body Serial Number"`
	FocalPlaneResolutionUnit       string `exifTag:"Focal Plane Resolution Unit"`
	Compression                    string `exifTag:"Compression"`
	CameraModel                    string `exifTag:"Model"`
	SubSecondTimeOriginal          string `exifTag:"Sub-second Time (Original)"`
	Software                       string `exifTag:"Software"`
	Saturation                     string `exifTag:"Saturation"`
	FocalLength                    string `exifTag:"Focal Length"`
	Flash                          string `exifTag:"Flash"`
	SensitivityType                string `exifTag:"Sensitivity Type"`
	LensSerialNumber               string `exifTag:"Lens Serial Number"`
	LensMake                       string `exifTag:"Lens Make"`
	LensSpecification              string `exifTag:"Lens Specification"`
	SubjectDistanceRange           string `exifTag:"Subject Distance Range"`
	Sharpness                      string `exifTag:"Sharpness"`
	FlashPixVersion                string `exifTag:"FlashPixVersion"`
	DateAndTimeDigitized           string `exifTag:"Date and Time (Digitized)"`
	DateAndTimeOriginal            string `exifTag:"Date and Time (Original)"`
	GainControl                    string `exifTag:"Gain Control"`
	SceneType                      string `exifTag:"Scene Type"`
	FocalPlaneYResolution          string `exifTag:"Focal Plane Y-Resolution"`
	ColorSpace                     string `exifTag:"Color Space"`
	XResolution                    string `exifTag:"X-Resolution"`
	LensModel                      string `exifTag:"Lens Model"`
	FocalLengthIn35mmFilm          string `exifTag:"Focal Length in 35mm Film"`
	CustomRendered                 string `exifTag:"Custom Rendered"`
	ExposureTime                   string `exifTag:"Exposure Time"`
	Aperture                       string `exifTag:"Aperture"`
	ExifVersion                    string `exifTag:"Exif Version"`
	SensingMethod                  string `exifTag:"Sensing Method"`
	LightSource                    string `exifTag:"Light Source"`
	ExposureBias                   string `exifTag:"Exposure Bias"`
	IsoSpeedRatings                string `exifTag:"ISO Speed Ratings"`
	YResolution                    string `exifTag:"Y-Resolution"`
	Contrast                       string `exifTag:"Contrast"`
	DateAndTime                    string `exifTag:"Date and Time"`
}

func ImageFromExif(img *Image, exifMap map[string]string) {
	s := reflect.ValueOf(img).Elem()
	t := s.Type()

	for i := 0; i < s.NumField(); i++ {
		field := t.Field(i)

		exifTagName := t.Field(i).Tag.Get(exifTag)
		exifTagValue := exifMap[exifTagName]

		if len(exifTagName) > 0 && len(exifTagValue) > 0 {
			value := reflect.ValueOf(img).Elem().FieldByName(field.Name)

			if value.IsValid() {
				value.SetString(exifTagValue)
			}
		}
	}

	fmt.Printf("%+v", img)
}
