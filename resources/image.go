package resources

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"gorm.io/gorm"
)

const exifTag = "exifTag"

type Image struct {
	FileId        string `gorm:"primarykey"`
	ImportBatchId string
	IsProcessed   bool
	WidthPixels   uint64
	HeightPixels  uint64

	// EXIF data
	Aperture                 float64   `exifTag:"Aperture"`
	ApertureValue            float64   `exifTag:"ApertureValue"`
	CameraModel              string    `exifTag:"Model"`
	ColorSpace               string    `exifTag:"ColorSpace"`
	DateTimeCreated          time.Time `exifTag:"DateTimeCreated"`
	DateTimeOriginal         time.Time `exifTag:"DateTimeOriginal"`
	DeviceManufacturer       string    `exifTag:"DeviceManufacturer"`
	DeviceModel              string    `exifTag:"DeviceModel"`
	DigitalCreationDateTime  time.Time `exifTag:"DigitalCreationDateTime"`
	ExposureCompensation     float64   `exifTag:"ExposureCompensation"`
	ExposureMode             string    `exifTag:"ExposureMode"`
	ExposureProgram          string    `exifTag:"ExposureProgram"`
	ExposureTime             string    `exifTag:"ExposureTime"`
	FileName                 string    `exifTag:"FileName"`
	Flash                    string    `exifTag:"Flash"`
	FNumber                  float64   `exifTag:"FNumber"`
	FocalLength              string    `exifTag:"FocalLength"`
	FocalLengthIn35mmFormat  string    `exifTag:"FocalLengthIn35mmFormat"`
	FocalPlaneResolutionUnit string    `exifTag:"FocalPlaneResolutionUnit"`
	FocalPlaneXResolution    float64   `exifTag:"FocalPlaneXResolution"`
	FocalPlaneYResolution    float64   `exifTag:"FocalPlaneYResolution"`
	Format                   string    `exifTag:"Format"`
	GPSAltitude              string    `exifTag:"GPSAltitude"`
	GPSDestBearing           string    `exifTag:"GPSDestBearing"`
	GPSImgDirection          string    `exifTag:"GPSImgDirection"`
	GPSLatitude              string    `exifTag:"GPSLatitude"`
	GPSLongitude             string    `exifTag:"GPSLongitude"`
	GPSPosition              string    `exifTag:"GPSPosition"`
	GPSSpeed                 string    `exifTag:"GPSSpeed"`
	ImageHeight              float64   `exifTag:"ImageHeight"`
	ImageNumber              float64   `exifTag:"ImageNumber"`
	ImageSize                string    `exifTag:"ImageSize"`
	ImageWidth               string    `exifTag:"ImageWidth"`
	ISO                      float64   `exifTag:"ISO"`
	Lens                     string    `exifTag:"Lens"`
	LensID                   string    `exifTag:"LensID"`
	LensInfo                 string    `exifTag:"LensInfo"`
	LensMake                 string    `exifTag:"LensMake"`
	LensModel                string    `exifTag:"LensModel"`
	LensSerialNumber         string    `exifTag:"LensSerialNumber"`
	Make                     string    `exifTag:"Make"`
	Megapixels               float64   `exifTag:"Megapixels"`
	MIMEType                 string    `exifTag:"MIMEType"`
	ModifyDate               string    `exifTag:"ModifyDate"`
	ResolutionUnit           string    `exifTag:"ResolutionUnit"`
	SerialNumber             string    `exifTag:"SerialNumber"`
	ShutterSpeed             string    `exifTag:"ShutterSpeed"`
	ShutterSpeedValue        string    `exifTag:"ShutterSpeedValue"`
	Software                 string    `exifTag:"Software"`
	XResolution              float64   `exifTag:"XResolution"`
	YResolution              float64   `exifTag:"YResolution"`
}

type ImportBatch struct {
	Id   string
	Date time.Time
}

type File struct {
	gorm.Model

	FileId        string // The uuid for the image
	ImportBatchId string // The id of the batch where the file was uploaded
	Filename      string // The filename with extension
	StoragePath   string // The path to the file, withing the storage bucket
	PublicURL     string // The public URL where the file is available
	IsOriginal    bool   // True if this is an original file
	Width         uint64 // Width in pixels of the image file
	Height        uint64 // Height in pixels of the image file
}

func parseExifTimestamp(s string) (time.Time, error) {
	// Exif date with timezone: 2021:12:11 09:17:18-08:00
	layout := "2006:01:02 15:04:05-07:00"
	value, err := time.Parse(layout, s)
	if err == nil {
		return value, nil
	}

	// Exif date format: 2021:12:10 20:29:21
	layout = "2006:01:02 15:04:05"
	value, err = time.Parse(layout, s)
	if err == nil {
		return value, nil
	}

	return time.Now(), err
}

func PopulateImageFromExif(img *Image, exifMap map[string]string) {
	s := reflect.ValueOf(img).Elem()
	t := s.Type()

	for i := 0; i < s.NumField(); i++ {
		field := t.Field(i)

		exifTagName := t.Field(i).Tag.Get(exifTag)
		exifTagValue := exifMap[exifTagName]

		if len(exifTagName) > 0 && len(exifTagValue) > 0 {
			value := reflect.ValueOf(img).Elem().FieldByName(field.Name)

			if value.IsValid() {
				if field.Type.Kind() == reflect.String {
					value.SetString(exifTagValue)
				}
				if field.Type.Kind() == reflect.Float64 {
					floatValue, err := strconv.ParseFloat(exifTagValue, 64)
					if err != nil {
						fmt.Printf("Unable to parse float %s\n", exifTagValue)
						continue
					}
					value.SetFloat(floatValue)
				}
				if field.Type.Kind() == reflect.Int64 {
					intValue, err := strconv.ParseInt(exifTagValue, 10, 64)
					if err != nil {
						fmt.Printf("Unable to parse int %s\n", exifTagValue)
						continue
					}
					value.SetInt(intValue)
				}
				// Assume structs are time.Time
				if field.Type.Kind() == reflect.Struct {
					timeValue, err := parseExifTimestamp(exifTagValue)

					if err != nil {
						fmt.Printf("Unable to parse date %s\n", exifTagValue)
						continue
					}

					value.Set(reflect.ValueOf(timeValue))
				}
			}
		}
	}
}
