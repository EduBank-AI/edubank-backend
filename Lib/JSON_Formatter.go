package Lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Struct for storing extracted text from PDF pages
type PageData struct {
	Page int    `json:"page"`
	Text string `json:"content"`
}

// Struct for PDF JSON format
type PDFExtracted struct {
	Pages []PageData `json:"pages"`
}

type Data struct {
	Topic   string      `json:"topic"`
	Content interface{} `json:"content"`
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
	var data []Data
	jsonFileName := "dataset.jsonl"

	_, err := os.Stat(jsonFileName)
	if os.IsNotExist(err) {
		createdFile, err := os.Create(jsonFileName)
		if err != nil {
			fmt.Println("Error making file: ", err)
			return err
		}
		defer createdFile.Close()

		_, err = createdFile.WriteString("[]")
		if err != nil {
			fmt.Println("Error writing file: ", err)
			return err
		}
	}

	file, err := ioutil.ReadFile(jsonFileName)
	if err != nil {
		fmt.Println("Error reading file: ", err)
		return err
	}

	err = json.Unmarshal(file, &data)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return err
	}

	newTopic := Data{
		Topic:   filename,
		Content: pages,
	}

	data = append(data, newTopic)
	updatedJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return err
	}

	err = ioutil.WriteFile(jsonFileName, updatedJSON, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return err
	}

	return nil
}

// Save extracted image text as JSONL (AI-friendly)
func saveImageAsJSONL(text string, filename string) error {
	var data []Data
	jsonFileName := "dataset.jsonl"

	_, err := os.Stat(jsonFileName)
	if os.IsNotExist(err) {
		createdFile, err := os.Create(jsonFileName)
		if err != nil {
			fmt.Println("Error making file: ", err)
			return err
		}
		defer createdFile.Close()

		_, err = createdFile.WriteString("[]")
		if err != nil {
			fmt.Println("Error writing file: ", err)
			return err
		}
	}

	file, err := ioutil.ReadFile(jsonFileName)
	if err != nil {
		fmt.Println("Error reading file: ", err)
		return err
	}

	err = json.Unmarshal(file, &data)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return err
	}

	newTopic := Data{
		Topic:   filename,
		Content: text,
	}

	data = append(data, newTopic)
	updatedJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return err
	}

	err = ioutil.WriteFile(jsonFileName, updatedJSON, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return err
	}

	return nil
}

func Format(pages []PageData, imageText string) error {
	// If pages are provided, process PDF
	if len(pages) > 0 {
		filename := extractFilename(pages[0].Text)
		err := savePDFAsJSON(pages, filename)

		if err != nil {
			fmt.Println("Error saving file:", err)
			return err
		}

		fmt.Println("Data of ", filename+" added to dataset.jsonl")
		return nil
	} else if imageText != "" {
		// If image text is provided, process image
		filename := extractFilename(imageText)
		cleanedText := cleanText(imageText)
		err := saveImageAsJSONL(cleanedText, filename)

		if err != nil {
			fmt.Println("Error saving file:", err)
			return err
		}
		
		fmt.Println("Data of ", filename+" added to dataset.jsonl")
		return nil
	}

	return errors.New("no valid input provided")
}
