package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/edubank/ai"
	"github.com/gin-gonic/gin"
)

// FileUploadHandler handles the /load POST request for file uploads
func FileUploadHandler(c *gin.Context) {
	// Receive the file
	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("Error retrieving file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Save file to Assets/
	dst := "ai/Assets/" + file.Filename
	if err := c.SaveUploadedFile(file, dst); err != nil {
		log.Printf("Error saving uploaded file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	log.Printf("File uploaded: %s", file.Filename)

	// Process the file
	if err := processFile(dst); err != nil {
		log.Printf("Error processing file %s: %v", file.Filename, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "File processing started",
		"filename": file.Filename,
	})
}

// processFile determines file type and calls AI library functions
func processFile(file string) error {
	if strings.HasSuffix(strings.ToLower(file), ".pdf") {
		log.Println("Starting PDF to text conversion...")
		extractedOutput, err := ai.PdfToText(file)
		if err != nil {
			return err
		}

		log.Println("Sending text to JSON formatter...")
		if err := ai.Format(extractedOutput, ""); err != nil {
			return err
		}

	} else if strings.HasSuffix(strings.ToLower(file), ".mp4") {
		log.Println("Starting video to text conversion...")
		audioFile := "ai/Assets/audio.wav"
		extractedOutput, err := ai.VidToText(file, audioFile)
		if err != nil {
			return err
		}

		log.Println("Sending text to JSON formatter...")
		if err := ai.Format(nil, extractedOutput); err != nil {
			return err
		}

	} else if strings.HasSuffix(strings.ToLower(file), ".png") ||
		strings.HasSuffix(strings.ToLower(file), ".jpeg") ||
		strings.HasSuffix(strings.ToLower(file), ".jpg") {

		log.Println("Starting image to text conversion...")
		extractedOutput, err := ai.ImgToText(nil, file)
		if err != nil {
			return err
		}

		log.Println("Sending text to JSON formatter...")
		if err := ai.Format(nil, extractedOutput); err != nil {
			return err
		}

	} else {
		return &UnsupportedFileTypeError{File: file}
	}

	return nil
}

// UnsupportedFileTypeError is returned when the uploaded file type is not supported
type UnsupportedFileTypeError struct {
	File string
}

func (e *UnsupportedFileTypeError) Error() string {
	return "unsupported file type: " + e.File
}
