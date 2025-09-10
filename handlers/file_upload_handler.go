package handlers

import (
	"log"
	"strings"

	"github.com/edubank/ai"
)

// FileUploadHandler handles the /load POST request for file uploads
func FileUploadHandler(dst string, jsonDir string) error {
	// Process the file
	if err := processFile(dst, jsonDir); err != nil {
		log.Printf("Error processing file %s: %v", dst, err)
		return err
	}

	return nil
}

// processFile determines file type and calls AI library functions
func processFile(file string, jsonFile string) error {
	if strings.HasSuffix(strings.ToLower(file), ".pdf") {
		log.Println("Starting PDF to text conversion...")
		extractedOutput, err := ai.PdfToText(file)
		if err != nil {
			return err
		}

		log.Println("Sending text to JSON formatter...")
		log.Println("Json File: ", jsonFile)
		if err := ai.Format(extractedOutput, "", jsonFile); err != nil {
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
		if err := ai.Format(nil, extractedOutput, jsonFile); err != nil {
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
		if err := ai.Format(nil, extractedOutput, jsonFile); err != nil {
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
