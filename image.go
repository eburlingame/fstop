package main

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"gorm.io/gorm"
)

const exifTag = "exifTag"

type Image struct {
	gorm.Model
	Filename string

	WidthPixels  uint32
	HeightPixels uint32

	// EXIF data
	BodySerialNumber      string `exifTag:"Body Serial Number"`
	CameraModel           string `exifTag:"Model"`
	Software              string `exifTag:"Software"`
	LensSerialNumber      string `exifTag:"Lens Serial Number"`
	LensMake              string `exifTag:"Lens Make"`
	LensSpecification     string `exifTag:"Lens Specification"`
	LensModel             string `exifTag:"Lens Model"`
	FocalLengthIn35mmFilm string `exifTag:"Focal Length in 35mm Film"`

	ShutterSpeed             string `exifTag:"Shutter Speed"`
	Manufacturer             string `exifTag:"Manufacturer"`
	CfaPattern               string `exifTag:"CFA Pattern"`
	MeteringMode             string `exifTag:"Metering Mode"`
	ExposureMode             string `exifTag:"Exposure Mode"`
	RecommendedExposureIndex string `exifTag:"Recommended Exposure Index"`
	FNumber                  string `exifTag:"F-Number"`
	SceneCaptureType         string `exifTag:"Scene Capture Type"`
	WhiteBalance             string `exifTag:"White Balance"`
	FileSource               string `exifTag:"File Source"`
	ExposureProgram          string `exifTag:"Exposure Program"`
	Compression              string `exifTag:"Compression"`
	Saturation               string `exifTag:"Saturation"`
	FocalLength              string `exifTag:"Focal Length"`
	Flash                    string `exifTag:"Flash"`
	SensitivityType          string `exifTag:"Sensitivity Type"`
	SubjectDistanceRange     string `exifTag:"Subject Distance Range"`
	Sharpness                string `exifTag:"Sharpness"`
	FlashPixVersion          string `exifTag:"FlashPixVersion"`
	GainControl              string `exifTag:"Gain Control"`
	SceneType                string `exifTag:"Scene Type"`
	ColorSpace               string `exifTag:"Color Space"`
	CustomRendered           string `exifTag:"Custom Rendered"`
	Aperture                 string `exifTag:"Aperture"`
	ExifVersion              string `exifTag:"Exif Version"`
	SensingMethod            string `exifTag:"Sensing Method"`
	LightSource              string `exifTag:"Light Source"`
	ExposureBias             string `exifTag:"Exposure Bias"`
	IsoSpeedRatings          string `exifTag:"ISO Speed Ratings"`
	Contrast                 string `exifTag:"Contrast"`
	ExposureTime             string `exifTag:"Exposure Time"`

	FocalPlaneXResolution    float32 `exifTag:"Focal Plane X-Resolution"`
	FocalPlaneYResolution    float32 `exifTag:"Focal Plane Y-Resolution"`
	FocalPlaneResolutionUnit string  `exifTag:"Focal Plane Resolution Unit"`
	XResolution              float32 `exifTag:"X-Resolution"`
	YResolution              float32 `exifTag:"Y-Resolution"`
	ResolutionUnit           string  `exifTag:"Resolution Unit"`

	OffsetTimeForDateTimeOriginal  string    `exifTag:"Offset Time For DateTimeOriginal"`
	OffsetTimeForDateTimeDigitized string    `exifTag:"Offset Time For DateTimeDigitized"`
	OffsetTimeForDateTime          string    `exifTag:"Offset Time For DateTime"`
	SubSecondTimeDigitized         string    `exifTag:"Sub-second Time (Digitized)"`
	SubSecondTimeOriginal          string    `exifTag:"Sub-second Time (Original)"`
	DateAndTimeDigitized           time.Time `exifTag:"Date and Time (Digitized)"`
	DateAndTimeOriginal            time.Time `exifTag:"Date and Time (Original)"`
	DateAndTime                    time.Time `exifTag:"Date and Time"`
}

func parseExifTimestamp(s string) (time.Time, error) {
	// Exif date format: 2021:12:10 20:29:21
	layout := "2006:01:02 15:04:05"

	return time.Parse(layout, s)
}

func PopulateImageFromExif(img *Image, exifMap map[string]string) {
	s := reflect.ValueOf(img).Elem()
	t := s.Type()

	for i := 0; i < s.NumField(); i++ {
		field := t.Field(i)

		exifTagName := t.Field(i).Tag.Get(exifTag)
		exifTagValue := exifMap[exifTagName]

		if len(exifTagName) > 0 && len(exifTagValue) > 0 {
			fmt.Printf("%s:\t\t\t%s\n", exifTagName, exifTagValue)

			value := reflect.ValueOf(img).Elem().FieldByName(field.Name)

			if value.IsValid() {
				if field.Type.Kind() == reflect.String {
					value.SetString(exifTagValue)
				}
				if field.Type.Kind() == reflect.Float32 {
					floatValue, err := strconv.ParseFloat(exifTagValue, 32)
					if err != nil {
						fmt.Printf("Unable to parse float %s", exifTagValue)
						continue
					}
					value.SetFloat(floatValue)
				}
				if field.Type.Kind() == reflect.Int32 {
					intValue, err := strconv.ParseInt(exifTagValue, 10, 32)
					if err != nil {
						fmt.Printf("Unable to parse int %s", exifTagValue)
						continue
					}
					value.SetInt(intValue)
				}
				// Assume structs are time.Time
				if field.Type.Kind() == reflect.Struct {
					timeValue, err := parseExifTimestamp(exifTagValue)

					if err != nil {
						fmt.Printf("Unable to parse date %s", exifTagValue)
						continue
					}

					value.Set(reflect.ValueOf(timeValue))
				}
			}
		}
	}
}
