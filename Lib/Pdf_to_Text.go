package Lib

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/google/generative-ai-go/genai"
	"github.com/unidoc/unipdf/v3/model"
	"github.com/unidoc/unipdf/v3/render"
	"google.golang.org/api/option"
)

func PdfToText(pdfFilePath string) {
	// Setup the directory where PNGs will be saved
	outputDir := "Assets/"
	outputDir = fmt.Sprintf("%s_images", pdfFilePath[:len(pdfFilePath)-len(".pdf")])
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, os.ModePerm)
	}

	// Step 1: Convert the pdf's pages to images
	fmt.Println("Converting the pdf's pages to images")
	extractPDFPagesAsImages(pdfFilePath, outputDir)

	finalText := ""

	// Step 2: Extract the text from each image
	fmt.Println("Extracting text from each image and sending it to gemini for cleanup")
	pageNo := 1
	err := filepath.WalkDir(outputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
			return err
		}
		if !d.IsDir() {
			extractedText, err := imgToText(path)
			if err != nil {
				log.Fatalf("Error in extracting tet from the images: %v", err)
			}

			// Step 3: Send the extracted text to gemini for cleanup
			finalText, err = pdfSendToGemini(extractedText)
			if err != nil {
				log.Fatalf("Error in sending text to gemini: %v", err)
			} else {
				fmt.Printf("Extracted text from page number %d: \n%s", pageNo, finalText)
				pageNo += 1
			}
		}
		fmt.Printf("\n\n\n")
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
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
	// Open the PDF file
	f, err := os.Open(pdfPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Read the PDF document
	doc, err := model.NewPdfReader(f)
	if err != nil {
		return err
	}

	// Iterate over each page
	numPages, err := doc.GetNumPages()
	if err != nil {
		return err
	}
	for i := 0; i < numPages; i++ {
		_, err := doc.GetPage(i + 1)
		if err != nil {
			return err
		}

		// Render the page to an image
		img, err := renderPDFPageToImage(doc, i+1)
		if err != nil {
			return err
		}

		// Save the image as PNG
		outputFilePath := fmt.Sprintf("%s/page_%d.png", outputDir, i+1)
		outFile, err := os.Create(outputFilePath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		err = png.Encode(outFile, img)
		if err != nil {
			return err
		}
	}

	return nil
}

// Placeholder function (requires PDF rendering library).
func renderPDFPageToImage(pdfReader *model.PdfReader, pageIndex int) (image.Image, error) {
	page, err := pdfReader.GetPage(pageIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get page %d: %w", pageIndex, err)
	}

	r := render.NewImageDevice()
	img, err := r.RenderWithOpts(page, false)
	if err != nil {
		return nil, fmt.Errorf("failed to render page %d: %w", pageIndex, err)
	}

	return img, nil
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
