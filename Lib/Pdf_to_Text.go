package Lib

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"io/fs"
	"os"
	"path/filepath"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func PdfToText(pdfFilePath string) ([]PageData, error) {
	// Setup the directory where PNGs will be saved
	outputDir := fmt.Sprintf("%s_images", pdfFilePath[:len(pdfFilePath)-len(".pdf")])
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, os.ModePerm)
	}

	// Step 1: Convert the pdf's pages to images
	fmt.Println("Converting the pdf's pages to images")
	extractPDFPagesAsImages(pdfFilePath, outputDir)

	finalText := ""

	// Step 2: Extract the text from each image
	fmt.Println("Extracting text from each image and sending it to gemini for cleanup")
	var pages []PageData
	pageNo := 1
	err := filepath.WalkDir(outputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			extractedText, err := imgToText(path)
			if err != nil {
				return err
			}

			// Step 3: Send the extracted text to gemini for cleanup
			finalText, err = pdfSendToGemini(extractedText)
			if err != nil {
				return err
			} else {
				pages = append(pages, PageData{Page: pageNo, Text: finalText})
				pageNo += 1
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return pages, nil
}

// Use gemini to cleanup the extrcated text
func pdfSendToGemini(text string) (string, error) {
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")

	// Create a new Gemini client using the API key
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("error creating client: %v", err)
	}
	defer client.Close()

	// Specify the model
	model := client.GenerativeModel("gemini-2.0-flash-001")

	// Create the prompt for summarization
	prompt := "Analyze the pdf and return the contents of the pdf in normal text format. Add double star for heading, single star for subheading, etc beautify the output a bit. Clean the text a bit like make the equations look good, etc. Read all the equations properly and solve them if unsolved. Do not summarize it and do not add etra texts like Here is the output, etc and show all the contents. If the formatted text is perfect then just return the text\n" + text

	// Generate the content
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("error generating content: %v", err)
	}

	// Check if the response contains candidates and extract the summarized text
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		summary, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
		if ok {
			// Return the summarized text
			return string(summary), nil
		} else {
			return "", fmt.Errorf("unexpected response format: could not extract text")
		}
	}

	// Return an error if no response content is found
	return "", fmt.Errorf("no response content found")
}

// Convert each page of a PDF into a PNG image.
func extractPDFPagesAsImages(pdfPath string, outputDir string) error {
	fmt.Println("Converting PDF to images...")
	cmd := exec.Command("pdftoppm", "-png", pdfPath, filepath.Join(outputDir, "page"))
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error converting PDF to images: %v\n", err)
		return err
	}

	return nil
}

// Extract the text from image
func imgToText(file string) (string, error) {

	ctx := context.Background()

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return "", err
	}

	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	image, err := vision.NewImageFromReader(f)
	if err != nil {
		return "", err
	}
	annotation, err := client.DetectDocumentText(ctx, image, nil)
	if err != nil {
		return "", err
	}

	if annotation == nil {
		return "", errors.New("no text found")
	} else {
		cleanOutput, err := imgSendToGemini(annotation.Text)
		if err != nil {
			return "", err
		} else {
			return cleanOutput, nil
		}
	}
}
