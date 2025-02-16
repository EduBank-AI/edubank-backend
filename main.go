package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const apiKey = "YOUR_GOOGLE_CLOUD_VISION_API_KEY" // Replace with your API Key

func main() {
	imagePath := "image.jpg"
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		log.Fatalf("Failed to read image: %v", err)
	}

	encodedImage := base64.StdEncoding.EncodeToString(imageData)

	requestBody, err := json.Marshal(map[string]interface{}{
		"requests": []map[string]interface{}{
			{
				"image": map[string]string{
					"content": encodedImage,
				},
				"features": []map[string]string{
					{"type": "TEXT_DETECTION"},
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create JSON request: %v", err)
	}

	url := fmt.Sprintf("https://vision.googleapis.com/v1/images:annotate?key=%s", apiKey)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response: %v", err)
	}

	fmt.Println("OCR Response:")
	fmt.Println(string(body))
}
