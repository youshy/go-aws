package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	AWSAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	AWSSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	S3Region := os.Getenv("S3_REGION")
	S3Bucket := os.Getenv("S3_BUCKET")

	log.Printf("Global variables:\nAccess key:\t%s\nSecret key:\t%s\nS3 Region:\t%s\nS3 Bucket:\t%s\n", AWSAccessKey, AWSSecretKey, S3Region, S3Bucket)
}

// init is invoked before main()
// by the design, if there's any new variable initialized here
// it'll be available wherever in the app
// init reads the .env file and sets up environmental variables
func init() {
	if err := godotenv.Load("config.env"); err != nil {
		log.Print("No .env file found")
	}
}
