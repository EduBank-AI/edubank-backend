package Lib

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Struct for storing extracted text from PDF pages
type PageData struct {
	Page int    `json:"page"`
	Text string `json:"text"`
}

// Struct for PDF JSON format
type PDFExtracted struct {
	Pages []PageData `json:"pages"`
}

// Struct for AI-friendly JSONL formatting
type AIFormatted struct {
	Text string `json:"text"`
}

// Function to extract the filename from the given text
func extractFilename(text string) string {
	lines := strings.Split(text, "\n")
	if len(lines) > 0 {
		// Take the first line and remove asterisks and extra spaces
		title := strings.TrimSpace(strings.TrimPrefix(lines[0], "**"))
		title = strings.TrimSpace(strings.TrimSuffix(title, "**"))
		title = strings.ReplaceAll(title, ":", "")  // Remove colons for filename safety
		title = strings.ReplaceAll(title, " ", "_") // Replace spaces with underscores
		return title
	}
	return "formatted_output" // Fallback filename
}

// Clean and normalize text for AI processing
func cleanText(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")     // Remove line breaks
	text = strings.ReplaceAll(text, "\t", " ")     // Remove tabs
	text = strings.Join(strings.Fields(text), " ") // Remove extra spaces
	return text
}

// Save extracted PDF text in a structured JSON
func savePDFAsJSON(pages []PageData, filename string) error {
	file, err := os.Create("Json/" + filename + ".json")
	if err != nil {
		return err
	}
	defer file.Close()

	data := PDFExtracted{Pages: pages}
	jsonData, _ := json.MarshalIndent(data, "", "  ")

	_, err = file.WriteString(string(jsonData))
	return err
}

// Save extracted image text as JSONL (AI-friendly)
func saveImageAsJSONL(text string, filename string) error {
	file, err := os.Create("Json/" + filename + ".jsonl")
	if err != nil {
		return err
	}
	defer file.Close()

	entry := AIFormatted{Text: text}
	jsonData, _ := json.Marshal(entry)

	_, err = file.WriteString(string(jsonData) + "\n")
	return err
}

func Format(pages []PageData, imageText string) {
	// If pages are provided, process PDF
	if len(pages) > 0 {
		filename := extractFilename(pages[0].Text)
		err := savePDFAsJSON(pages, filename)

		if err != nil {
			fmt.Println("Error saving file:", err)
		} else {
			fmt.Println("File saved as:", filename+".jsonl")
		}
	} else if imageText != "" {
		// If image text is provided, process image
		filename := extractFilename(imageText)
		cleanedText := cleanText(imageText)
		err := saveImageAsJSONL(cleanedText, filename)

		if err != nil {
			fmt.Println("Error saving file:", err)
		} else {
			fmt.Println("File saved as:", filename+".jsonl")
		}
	} else {
		fmt.Println("no valid input provided")
	}
}
