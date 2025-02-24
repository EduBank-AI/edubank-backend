package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/edubank/Lib"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// file := "Assets/video.mp4"
	// file := "Assets/image.png"
	file := "Assets/5.2 The Definite Integral (and review of Riemann sums).pdf"

	if strings.HasSuffix(strings.ToLower(file), ".pdf") {
		fmt.Println("Starting PDF to text conversion...")
		extractedOutput, err := Lib.PdfToText(file)

		if err != nil {
			fmt.Println("Error extracting pdf text")
		}

		fmt.Println("Sending text to JSON formatter")
		Lib.Format(extractedOutput, "")
	} else if strings.HasSuffix(strings.ToLower(file), ".mp4") {
		audioFile := "Assets/audio.wav"

		fmt.Println("Starting video to text conversion...")
		extractedOutput, err := Lib.VidToText(file, audioFile)

		if err != nil {
			fmt.Println("Error extracting pdf text")
		}

		fmt.Println("Sending text to JSON formatter")
		Lib.Format(nil, extractedOutput)
	} else if strings.HasSuffix(strings.ToLower(file), ".png") || strings.HasSuffix(strings.ToLower(file), ".jpeg") || strings.HasSuffix(strings.ToLower(file), ".jpg") {
		fmt.Println("Starting image to text conversion...")
		extractedOutput, err := Lib.ImgToText(os.Stdout, file)

		if err != nil {
			fmt.Println("Error extracting image text")
		}

		fmt.Println("Sending text to JSON formatter")
		Lib.Format(nil, extractedOutput)
	}
}
