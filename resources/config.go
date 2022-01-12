package resources

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type Configuration struct {
	Secret string
	ApiKey string

	SQLiteFilepath string

	AdminUsername     string
	AdminPasswordHash []byte

	S3BucketName   string
	S3BucketRegion string
	S3MediaFolder  string
	S3UploadFolder string
	S3BaseUrl      string
}

func GetConfig() *Configuration {
	if err := godotenv.Load(".env.local"); err != nil {
		log.Print(err)
	}

	password := os.Getenv("ADMIN_PASSWORD")
	password_bytes := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(password_bytes, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	return &Configuration{
		Secret:         os.Getenv("SECRET"),
		ApiKey:         os.Getenv("API_KEY"),
		SQLiteFilepath: os.Getenv("SQLITE_FILE"),

		AdminUsername:     os.Getenv("ADMIN_USERNAME"),
		AdminPasswordHash: hashedPassword,

		S3BucketName:   os.Getenv("S3_BUCKET_NAME"),
		S3BucketRegion: os.Getenv("S3_BUCKET_REGION"),
		S3MediaFolder:  os.Getenv("S3_BUCKET_MEDIA_FOLDER"),
		S3UploadFolder: os.Getenv("S3_BUCKET_UPLOAD_FOLDER"),
		S3BaseUrl:      os.Getenv("S3_BUCKET_PUBLIC_BASE_URL"),
	}
}
