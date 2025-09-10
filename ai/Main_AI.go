package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// =============== Structs ===============
type Content struct {
	Content string `json:"content"`
	Page    int    `json:"page,omitempty"`
}

type Topic struct {
	Topic   string    `json:"topic"`
	Content []Content `json:"content,omitempty"`
}

// For topics that sometimes use string instead of array
func (t *Topic) UnmarshalJSON(data []byte) error {
	type Alias Topic
	var tmp Alias
	if err := json.Unmarshal(data, &tmp); err == nil {
		*t = Topic(tmp)
		return nil
	}

	var alt struct {
		Topic   string `json:"topic"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(data, &alt); err != nil {
		return err
	}
	*t = Topic{
		Topic: alt.Topic,
		Content: []Content{
			{Content: alt.Content},
		},
	}
	return nil
}

// =============== Core AI System ===============
type QASystem struct {
	topics []Topic
	apiKey string
}

// NewQASystem loads dataset into memory
func NewQASystem(datasetPath, apiKey string) (*QASystem, error) {
	data, err := os.ReadFile(datasetPath)
	if err != nil {
		return nil, fmt.Errorf("error reading dataset: %v", err)
	}

	var topics []Topic
	if err := json.Unmarshal(data, &topics); err != nil {
		return nil, fmt.Errorf("error parsing dataset: %v", err)
	}

	return &QASystem{topics: topics, apiKey: apiKey}, nil
}

// FindRelevantContent tries to match question to dataset topics
func (qa *QASystem) FindRelevantContent(question string) []string {
	questionLower := strings.ToLower(question)
	var results []string

	for _, topic := range qa.topics {
		topicLower := strings.ToLower(strings.ReplaceAll(topic.Topic, "_", " "))

		if strings.Contains(questionLower, topicLower) ||
			strings.Contains(topicLower, questionLower) {
			for _, c := range topic.Content {
				results = append(results, c.Content)
			}
		}
	}
	return results
}

// queryGemini calls Gemini API
func (qa *QASystem) queryGemini(prompt string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(qa.apiKey))
	if err != nil {
		return "", err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash-001")

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			return string(text), nil
		}
	}
	return "", fmt.Errorf("empty response from Gemini")
}

// =============== Public Entry ===============

// AI is the single entrypoint for handlers
// mode = "qa" | "exam" | "transform"
func AI(mode, question, datasetPath string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not set")
	}

	qa, err := NewQASystem(datasetPath, apiKey)
	if err != nil {
		return "", err
	}

	contextData := qa.FindRelevantContent(question)
	contextStr := strings.Join(contextData, "\n\n")

	var prompt string
	switch mode {
	case "qa":
		prompt = fmt.Sprintf(
			"You are a helpful assistant. Answer ONLY from context.\n"+
				"If no info is found, reply: 'I don't have enough information to answer that question.'\n\n"+
				"Context:\n%s\n\nQuestion: %s",
			contextStr, question)

	case "exam":
		// Example: question = "topic=Work, count=5, difficulty=medium"
		prompt = fmt.Sprintf(
			"You are an exam question generator.\n"+
				"Using the provided dataset, generate unique questions along with the answers.\n"+
				"For each question: \n" +
				"Make sure it is relevant to the topic and difficulty specified. \n" +
				"Provide a clear, correct answer immediately after the question. \n" +
				"Do not include any additional explanation or commentary. \n" +
				"%s\n\nContext:\n%s"+
				"Format: \n" +
				"**Question <question number>:** \n" +
				"<Question Text> \n" +
				"**Answer <answer number>:** \n" +
				"<Answer Text>",
			question, contextStr)

	case "transform":
		// Example: question = "Evaluate the integral: ∫ 6x² cos(2x³+1) dx."
		prompt = fmt.Sprintf(
			"You are a word problem transformer.\n"+
				"Take the given problem: %s\n"+
				"1. Generate a new version of the question by only changing the numeric values (slightly). \n" +
				"2. Preserve the logical structure of the question. \n" + 
				"3. Then, compute and display the correct answer to the new question. \n" +
				"4. Only output the transformed question and its answer. \n" +
				"Format: \n" +
				"Question: \n" +
				"<your transformed version of the question> \n" +
				"Answer: \n" +
				"<correct answer to the transformed question>",
			question)

	default:
		return "", fmt.Errorf("invalid mode: %s", mode)
	}

	return qa.queryGemini(prompt)
}
