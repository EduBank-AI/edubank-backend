package handlers

import (
	"log"
	"net/http"

	"github.com/edubank/ai"
	"github.com/gin-gonic/gin"
)

// AIHandler handles the /ai POST request
func AIHandler(c *gin.Context) {
	log.Println("Received AI request")

	var request struct {
		Question string `json:"question"`
	}

	if err := c.BindJSON(&request); err != nil {
		log.Printf("Invalid AI request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Call AI function
	answer := ai.AI(request.Question)

	log.Printf("Question: %s | Answer: %s", request.Question, answer)
	c.JSON(http.StatusOK, gin.H{"answer": answer})
}
