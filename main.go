package main

import (
	"bytes"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
	"gopkg.in/mgo.v2/bson"
)

func main() {
	http.HandleFunc("/", handler)
	log.Println("Upload server started")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// init is invoked before main() by design
// init reads the .env file and sets up environmental variables
func init() {
	if err := godotenv.Load("config.env"); err != nil {
		log.Print("No .env file found")
	}
}

// UploadFileToS3 saves a file to aws bucket and returns the url to the file and an error (if there's any)
func UploadFileToS3(s *session.Session, file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	// env variables
	S3Bucket := os.Getenv("S3_BUCKET")

	// get the file size and read
	// the file content into a buffer
	size := fileHeader.Size
	buffer := make([]byte, size)
	file.Read(buffer)

	// create a unique file name for the file
	tempFileName := "pictures/" + bson.NewObjectId().Hex() + filepath.Ext(fileHeader.Filename)

	log.Printf("tempfilename: %s\n", tempFileName)
	// Config settings: this sets up for uploaded file:
	// - bucket
	// - filename
	// - content-type
	// - file's storage class
	// establishes a new session to S3 and saves the object
	_, err := s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(S3Bucket),
		Key:                  aws.String(tempFileName),
		ACL:                  aws.String("public-read"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(int64(size)),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
		StorageClass:         aws.String("INTELLIGENT_TIERING"),
	})
	if err != nil {
		return "", err
	}
	log.Printf("Uploaded to S3: %s", tempFileName)
	return tempFileName, err
}

// handler handles the save function for image
func handler(w http.ResponseWriter, r *http.Request) {
	// env variables
	AWSAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	AWSSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	S3Region := os.Getenv("S3_REGION")

	maxSize := int64(2048000) // allows for max 2MB of file size

	err := r.ParseMultipartForm(maxSize) // checks if the form is below the max size
	if err != nil {
		log.Println(err)
		fmt.Fprintf(w, "Image too large. Max Size: %v", maxSize)
		return
	}
	log.Printf("Req: %v\n", r)

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		fmt.Fprintf(w, "Could not get uploaded file")
		return
	}
	defer file.Close()

	// create an AWS session which can be
	// reused if uploading many files

	s, err := session.NewSession(&aws.Config{
		Region: aws.String(S3Region),
		Credentials: credentials.NewStaticCredentials(
			AWSAccessKey, // id
			AWSSecretKey, // secret
			""),          // token, can be blank
	})
	if err != nil {
		fmt.Fprint(w, "Could not establish connection to S3")
		return
	}
	fmt.Println("Established connection to S3")

	fileName, err := UploadFileToS3(s, file, fileHeader)
	if err != nil {
		fmt.Fprintf(w, "Could not upload file to S3")
		return
	}

	fmt.Fprintf(w, "Image uploaded successfully: %v", fileName)
}
