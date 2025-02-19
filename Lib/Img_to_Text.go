package Lib

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	vision "cloud.google.com/go/vision/apiv1"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// detectDocumentText gets the full document text from the Vision API for an image at the given file path.
func ImgToText(w io.Writer, file string) {
	ctx := context.Background()

	fmt.Println("Extracting text from the image")

	// Define the vision client
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		log.Fatalf("Error defining the vision client: %v", err)
	}

	// Open the image file
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("Error opening the image file: %v", err)
	}
	defer f.Close()

	// Step 1: Use vision to read the image and extract the text
	image, err := vision.NewImageFromReader(f)
	if err != nil {
		log.Fatalf("Error reading the image: %v", err)
	}
	annotation, err := client.DetectDocumentText(ctx, image, nil)
	if err != nil {
		log.Fatalf("Error detecting the tet: %v", err)
	}

	// Check if the extracted text is empty or not
	if annotation == nil {
		fmt.Fprintln(w, "No text found.")
	} else {
		// Step 2: Send the text to gemini for cleanup
		fmt.Println("Sending the text to Gemini for cleanup")
		cleanOutput, err := imgSendToGemini(annotation.Text)
		if err != nil {
			log.Fatalf("Error in summarizing the text: %v", err)
		} else {
			fmt.Fprintf(w, "Extracted text: \n%s", cleanOutput)
		}
	}
}

// Use gemini to cleanup the extrcated text
func imgSendToGemini(text string) (string, error) {
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
	prompt := "Analyze the text contents. Clean the text a bit like make the equations look good, etc. Do not summarize it and show all the contents\n" + text

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
