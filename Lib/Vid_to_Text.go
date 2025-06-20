package Lib

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	speech "cloud.google.com/go/speech/apiv1"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

func VidToText(videoFile string, audioFile string) (string, error) {

	// Step 1: Extract audio from video
	fmt.Println("Extracting audio...")
	err := extractAudio(videoFile, audioFile)
	if err != nil {
		log.Fatalf("Audio extraction failed: %v", err)
	}

	// Step 2: Transcribe audio to text
	fmt.Println("Transcribing audio...")
	transcribedText, err := transcribeAudio(audioFile)
	if err != nil {
		return "", err
	}

	// Step 3: Send transcribed text to Gemini API for summarization
	fmt.Println("Sending text to Gemini API for summarization...")
	summarizedText, err := vidSendToGemini(transcribedText)
	if err != nil {
		return "", err
	}

	// Output the transcribed text and summarized text
	// fmt.Println("\n--- Transcribed Text ---")

	return summarizedText, nil
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
	audioData, err := ioutil.ReadFile(audioPath)
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

// Sends transcribed text to Gemini API for summarizationṇ
func vidSendToGemini(text string) (string, error) {
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
	prompt := "Analyze the text and return more logical version of the text also clean it a bit. Add double star for heading, single star for subheading, etc beautify the output a bit. Do not summarize it. And also do not show anything else other than the text. If the formatted text is perfect then just return the text\n" + text

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
