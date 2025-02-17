package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	speech "cloud.google.com/go/speech/apiv1"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	videoFile := "video.mp4"
	audioFile := "audio.wav"

	// Step 1: Extract audio from video
	fmt.Println("Extracting audio...")
	err = extractAudio(videoFile, audioFile)
	if err != nil {
		log.Fatalf("Audio extraction failed: %v", err)
	}
	fmt.Println("Audio extracted successfully!")

	// Step 2: Transcribe audio to text
	fmt.Println("Transcribing audio...")
	transcribedText, err := transcribeAudio(audioFile)
	if err != nil {
		log.Fatalf("Transcription failed: %v", err)
	}
	fmt.Println("Transcription complete!")

	// Step 3: Send transcribed text to Gemini API for summarization
	fmt.Println("Sending text to Gemini API for summarization...")
	summarizedText, err := sendToGemini(transcribedText)
	if err != nil {
		log.Fatalf("Gemini API call failed: %v", err)
	}

	// Output the transcribed text and summarized text
	fmt.Println("\n--- Transcribed Text ---")
	fmt.Println(transcribedText)

	fmt.Println("\n--- Gemini Response (Summarized Text) ---")
	fmt.Println(summarizedText)
}

// Extracts audio from the video using ffmpeg and converts it to mono
func extractAudio(videoPath, audioPath string) error {
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-vn", "-ac", "1", "-ar", "16000", "-acodec", "pcm_s16le", audioPath, "-y")
	return cmd.Run()
}

// Transcribes audio to text using Google Speech-to-Text API
func transcribeAudio(audioPath string) (string, error) {
	ctx := context.Background()
	client, err := speech.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	// Read the audio file
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return "", err
	}

	// Prepare the request
	req := &speechpb.RecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:     speechpb.RecognitionConfig_LINEAR16, // Assuming audio is in LINEAR16 format
			LanguageCode: "en-US",
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{Content: audioData},
		},
	}

	// Call Speech-to-Text API
	resp, err := client.Recognize(ctx, req)
	if err != nil {
		return "", err
	}

	// Collect transcribed text
	var transcript string
	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			transcript += alt.Transcript + " "
		}
	}
	return transcript, nil
}

// Sends transcribed text to Gemini API for summarization
func sendToGemini(text string) (string, error) {
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")

	// Create a new Gemini client using the API key
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return "", fmt.Errorf("error creating client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash-001")

	// Summarisation prompt
	prompt := "Analyse the text and create a question bank of 3 multiple choice questions with one option correct along with their answers from different topics if possible. I want the output you give me to be in a specifc format therefore do not include nay other lines of text or filler sentences before giving the main output. Question<Question Number>(next line)<Question><All options labeled a,b,c,d>(Next Line)Answer<Answer Number>(next line)<Answer>\n" + text

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

	return "", fmt.Errorf("no response content found")
}
