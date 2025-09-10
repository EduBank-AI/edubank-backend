package handlers

import (
	"log"
	"net/http"
	"fmt"
	"os"

	"github.com/edubank/ai"
	"github.com/edubank/db"
	"github.com/gin-gonic/gin"
)

// AIHandler handles the /ai POST request
func AIHandler(c *gin.Context) {
	log.Println("Received AI request")

	email := c.GetString("email")
	ctx := c.Request.Context()

	// Get user ID from email
	var userID int
	if err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE email=$1", email).Scan(&userID); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	// Construct dataset path
	datasetPath := fmt.Sprintf("ai/users/%d/dataset.jsonl", userID)

	// Check if file exists
	if _, err := os.Stat(datasetPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "dataset not found"})
		return
	}

	// Bind incoming JSON request
	var request struct {
		Question string `json:"question"`
		Mode     string `json:"mode"` // "qa", "exam", "transform"
	}

	if err := c.BindJSON(&request); err != nil {
		log.Printf("Invalid AI request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if request.Mode == "" {
		request.Mode = "qa" // default to normal QA
	}

	// Call AI function
	answer, err := ai.AI(request.Mode, request.Question, datasetPath)
	if err != nil {
		log.Printf("Error processing AI request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("User: %s | Mode: %s | Question: %s | Answer: %s", email, request.Mode, request.Question, answer)
	c.JSON(http.StatusOK, gin.H{"answer": answer})
}
