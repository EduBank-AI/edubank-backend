package Lib

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Define structs that match your JSON structure
type Content struct {
	Content string `json:"content"`
	Page    int    `json:"page,omitempty"`
}

type Topic struct {
	Topic   string    `json:"topic"`
	Content []Content `json:"content,omitempty"`
}

// For topics with a single content string
func (t *Topic) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as the standard structure
	type TopicAlias Topic
	var standard TopicAlias
	if err := json.Unmarshal(data, &standard); err == nil {
		*t = Topic(standard)
		return nil
	}

	// If that fails, try to unmarshal with string content
	var alternative struct {
		Topic   string `json:"topic"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(data, &alternative); err != nil {
		return err
	}

	// Convert string content to Content slice
	*t = Topic{
		Topic: alternative.Topic,
		Content: []Content{
			{Content: alternative.Content},
		},
	}
	return nil
}

// QASystem represents our question answering system
type QASystem struct {
	topics      []Topic
	apiKey      string
	apiEndpoint string
	modelName   string
}

// AI API request/response structures
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewQASystem creates a new QA system from a JSON file
func NewQASystem(jsonFilePath, apiKey string) (*QASystem, error) {
	// Read the JSON file
	data, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Parse the JSON data
	var topics []Topic
	if err := json.Unmarshal(data, &topics); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return &QASystem{
		topics:      topics,
		apiKey:      apiKey,
		apiEndpoint: "https://api.openai.com/v1/chat/completions",
		modelName:   "gpt-3.5-turbo", // You can change this to a different model
	}, nil
}

// FindRelevantContent searches for content relevant to the question
func (qa *QASystem) FindRelevantContent(question string) []string {
	// Convert question to lowercase for case-insensitive matching
	questionLower := strings.ToLower(question)

	// Search for relevant content
	var relevantResults []string

	// First, look for topic matches
	for _, topic := range qa.topics {
		topicNormalized := strings.ToLower(strings.Replace(topic.Topic, "_", " ", -1))

		// Check if the topic is relevant to the question
		if strings.Contains(questionLower, topicNormalized) ||
			strings.Contains(questionLower, strings.ToLower(topic.Topic)) {

			// If we found a relevant topic, extract all content
			for _, content := range topic.Content {
				relevantResults = append(relevantResults, content.Content)
			}
		}
	}

	// If no topic matches, try content keyword matching
	if len(relevantResults) == 0 {
		// Extract keywords from the question (simple approach)
		keywords := extractKeywords(questionLower)

		for _, topic := range qa.topics {
			for _, content := range topic.Content {
				contentLower := strings.ToLower(content.Content)

				// Check if content contains any keywords
				for _, keyword := range keywords {
					if strings.Contains(contentLower, keyword) {
						relevantResults = append(relevantResults, content.Content)
						break // Only add each content once
					}
				}
			}
		}
	}

	return relevantResults
}

// QueryAI sends a question to the AI API with relevant context
// QueryAI sends a question to the Gemini API with relevant context
func (qa *QASystem) QueryAI(question string, contextData []string) (string, error) {
	// Combine context into a single string
	contextStr := strings.Join(contextData, "\n\n")

	// Create context for the API call
	ctx := context.Background()

	// Initialize the Gemini client
	client, err := genai.NewClient(ctx, option.WithAPIKey(qa.apiKey))
	if err != nil {
		return "", fmt.Errorf("error creating Gemini client: %v", err)
	}
	defer client.Close()

	// Initialize the model
	model := client.GenerativeModel("gemini-2.0-flash-001")

	// Create prompt with context and question
	prompt := fmt.Sprintf("You are a helpful assistant that answers questions based ONLY on the provided context. "+
		"If the answer cannot be found in the context, say 'I don't have enough information to answer that question.' "+
		"Do not use any knowledge outside of the provided context.\n\n"+
		"Context:\n%s\n\nQuestion: %s\n\nAnswer the question based only on the provided context.",
		contextStr, question)

	// Generate content
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("error generating content: %v", err)
	}

	// Extract response text
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if textResp, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			return string(textResp), nil
		}
	}

	return "", fmt.Errorf("no response from Gemini service")
}

// AnswerQuestion attempts to answer a question based on the JSON content
func (qa *QASystem) AnswerQuestion(question string) string {
	// Find relevant content
	relevantContent := qa.FindRelevantContent(question)

	// If we found relevant content, query the AI
	if len(relevantContent) > 0 {
		answer, err := qa.QueryAI(question, relevantContent)
		if err != nil {
			return fmt.Sprintf("Error querying AI: %v\n\nFallback answer based on document:\n\n%s",
				err, strings.Join(relevantContent, "\n\n"))
		}
		return answer
	}

	return "I couldn't find information related to your question in the document."
}

// Simple keyword extraction (you may want to improve this)
func extractKeywords(question string) []string {
	// Remove common words and punctuation
	stopWords := map[string]bool{
		"what": true, "where": true, "when": true, "how": true, "why": true,
		"is": true, "are": true, "the": true, "a": true, "an": true, "in": true,
		"of": true, "to": true, "for": true, "on": true, "with": true, "about": true,
	}

	words := strings.Fields(question)
	var keywords []string

	for _, word := range words {
		// Clean the word
		word = strings.Trim(word, ".,?!;:\"'()[]{}")
		word = strings.ToLower(word)

		// Skip if it's empty or a stop word
		if word == "" || stopWords[word] {
			continue
		}

		keywords = append(keywords, word)
	}

	return keywords
}

func AI(question string) string {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: GEMINI_API_KEY environment variable is not set")
		os.Exit(1)
	}

	// Create a new QA system
	qa, err := NewQASystem("dataset.jsonl", apiKey)
	if err != nil {
		fmt.Printf("Error initializing QA system: %v\n", err)
		os.Exit(1)
	}

	answer := qa.AnswerQuestion(question)
	return answer
}
