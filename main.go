package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
)

const (
	projectID  = "myits-student-registration"
	bucketName = "misr-sipmaba-bucket" // FILL IN WITH YOURS
)

type ClientUploader struct {
	cl         *storage.Client
	projectID  string
	bucketName string
	uploadPath string
}

var uploader *ClientUploader

func init() {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "myits-student-registration-913d13ea49bc.json") // FILL IN WITH YOUR FILE PATH
	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	uploader = &ClientUploader{
		cl:         client,
		bucketName: bucketName,
		projectID:  projectID,
		uploadPath: "test-files/",
	}

}

func main() {
	r := gin.Default()
	r.POST("/upload", func(c *gin.Context) {
		f, err := c.FormFile("file_input")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		blobFile, err := f.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		err = uploader.UploadFile(blobFile, f.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "success",
		})
	})

	r.Run()
}

// UploadFile uploads an object
func (c *ClientUploader) UploadFile(file multipart.File, object string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := c.cl.Bucket(c.bucketName).Object(c.uploadPath + object).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}
