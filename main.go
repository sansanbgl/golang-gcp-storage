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
	"github.com/google/uuid"
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

		fileID, err := uploader.UploadFile(blobFile, f.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(200, gin.H{"message": "success", "fileID": fileID})
	})
	r.GET("/download/:filename", func(c *gin.Context) {
		fileName := c.Param("filename")

		// Retrieve the file from GCS using the ClientUploader
		reader, err := uploader.DownloadFile(fileName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to download file: " + err.Error(),
			})
			return
		}
		// Set content disposition to suggest a filename for download
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
		// Stream the file content directly to the response
		c.Stream(func(w io.Writer) bool {
			defer reader.Close() // Close the reader when finished

			// Copy the file content to the response body
			_, err := io.Copy(w, reader)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to stream file: " + err.Error(),
				})
				return false // Stop streaming
			}
			return true // Indicate successful streaming
		})
	})
	r.Run()
}

// UploadFile uploads an object
func (c *ClientUploader) UploadFile(file multipart.File, object string) (string, error) {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	objectName := fmt.Sprintf("%s", uuid.New().String())
	oName := objectName + object
	wc := c.cl.Bucket(c.bucketName).Object(c.uploadPath + oName).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return "", fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("Writer.Close: %v", err)
	}
	return oName, nil
}

func (c *ClientUploader) DownloadFile(fileName string) (io.ReadCloser, error) {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	// Get a reference to the object with the filename
	reader, err := c.cl.Bucket(c.bucketName).Object(c.uploadPath + fileName).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewReader: %v", err)
	}
	return reader, nil
}

// misr-sipmaba-bucket/test-files/WhatsApp Image 2024-02-12 at 10.59.51_abe48a3c.jpg-5870d01c-b0f4-48ab-9e43-ec0033e9e6fc
