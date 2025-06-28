package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/edubank/Lib"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.POST("/load", handleFileLoad)
	r.POST("/ai", handleAIRequest) // Add AI route

	return r
}

func handleFileLoad(c *gin.Context) {
	// Receive the file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "No file uploaded"})
		return
	}

	// Save file to a temporary location
	dst := "Assets/" + file.Filename
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}

	// Use your existing loadData function to process the file
	if err := loadData(dst); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Respond to the frontend
	c.JSON(200, gin.H{
		"message":  "File processing started",
		"filename": file.Filename,
	})
}

func handleAIRequest(c *gin.Context) {
	var request struct {
		Question string `json:"question"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Call AI function
	answer := Lib.AI(request.Question)

	c.JSON(200, gin.H{"answer": answer})
}

func loadData(file string) error {
	if strings.HasSuffix(strings.ToLower(file), ".pdf") {
		fmt.Println("Starting PDF to text conversion...")
		extractedOutput, err := Lib.PdfToText(file)

		if err != nil {
			fmt.Println("Error extracting pdf text: ", err)
			return fmt.Errorf("pdf extraction failed: %v", err)
		}

		fmt.Println("Sending text to JSON formatter")

		lib_err := Lib.Format(extractedOutput, "")
		if lib_err != nil {
			fmt.Println("Error saving datat to jsonl: ", lib_err)
			return fmt.Errorf("error in jsonl formatting: %v", lib_err)
		}
	} else if strings.HasSuffix(strings.ToLower(file), ".mp4") {
		audioFile := "Assets/audio.wav"

		fmt.Println("Starting video to text conversion...")
		extractedOutput, err := Lib.VidToText(file, audioFile)

		if err != nil {
			fmt.Println("Error extracting video text")
			return fmt.Errorf("video transcription failed")
		}

		fmt.Println("Sending text to JSON formatter")

		lib_err := Lib.Format(nil, extractedOutput)
		if lib_err != nil {
			fmt.Println("Error saving datat to jsonl: ", lib_err)
			return fmt.Errorf("error in jsonl formatting: %v", lib_err)
		}
	} else if strings.HasSuffix(strings.ToLower(file), ".png") || strings.HasSuffix(strings.ToLower(file), ".jpeg") || strings.HasSuffix(strings.ToLower(file), ".jpg") {
		fmt.Println("Starting image to text conversion...")
		extractedOutput, err := Lib.ImgToText(os.Stdout, file)

		if err != nil {
			fmt.Println("Error extracting image text: ", err)
			return fmt.Errorf("image ocr failed: %v", err)
		}

		fmt.Println("Sending text to JSON formatter")

		lib_err := Lib.Format(nil, extractedOutput)
		if lib_err != nil {
			fmt.Println("Error saving datat to jsonl: ", lib_err)
			return fmt.Errorf("error in jsonl formatting: %v", lib_err)
		}

	} else {
		return fmt.Errorf("unsupported file type")
	}
	return nil
}

func ai(question string) {
	// fmt.Println("\nAsk a question:")
	// reader := bufio.NewReader(os.Stdin)
	// question, _ := reader.ReadString('\n')
	// question = strings.TrimSpace(question)

	fmt.Println("Processing your question...")
	answer := Lib.AI(question)
	fmt.Println("\nAnswer:")
	fmt.Println(answer)
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Start the server
	r := setupRouter()
	port := ":5000" // Change this if needed
	fmt.Println("Server running on http://localhost" + port)
	if err := r.Run(port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
