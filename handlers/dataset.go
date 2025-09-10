package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"log"

	"github.com/edubank/db"
	"github.com/gin-gonic/gin"
)

// UploadDatasetHandler
func UploadDatasetHandler(c *gin.Context) {
	email := c.GetString("email") // from AuthMiddleware
	ctx := context.Background()

	// Get user ID from email
	var userID int
	if err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE email=$1", email).Scan(&userID); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	// Get uploaded file
	file, err := c.FormFile("dataset")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file"})
		return
	}

	// Create user-specific directory
	userDir := fmt.Sprintf("ai/users/%d", userID)

	// Create Assets directory inside user dir
	assetsDir := fmt.Sprintf("%s/Assets", userDir)
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user dir"})
		return
	}

	savePath := filepath.Join(assetsDir, file.Filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file save failed"})
		return
	}

	// Save metadata in database
	var existingID int
	err = db.Pool.QueryRow(ctx,
	    "SELECT id FROM datasets WHERE filename=$1 AND user_id=$2",
	    file.Filename, userID,
	).Scan(&existingID)

	if err == nil {
        // Replace existing record
        _, err := db.Pool.Exec(ctx,
			"UPDATE datasets SET file_url=$1, uploaded_at=$2 WHERE filename=$3 AND user_id=$4", savePath, time.Now(), file.Filename, userID,
		)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "db update failed"})
			return
        }
    } else {
        // Insert new record
        _, err := db.Pool.Exec(ctx,
			"INSERT INTO datasets (user_id, filename, file_url, size_bytes, uploaded_at) VALUES ($1,$2,$3,$4,$5)",
			userID, file.Filename, savePath, file.Size, time.Now(),
		)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed"})
			return
        }
    }

	log.Printf("Dataset path: %s/dataset.jsonl", userDir)
	if err := FileUploadHandler(savePath, fmt.Sprintf("%s/dataset.jsonl", userDir)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file processing failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "dataset uploaded", "filename": file.Filename})
}


// ListDatasetsHandler
func ListDatasetsHandler(c *gin.Context) {
	email := c.GetString("email")
	ctx := context.Background()

	var userID int
	if err := db.Pool.QueryRow(ctx, "SELECT id FROM users WHERE email=$1", email).Scan(&userID); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	rows, err := db.Pool.Query(ctx,
		"SELECT id, filename, file_url, size_bytes, uploaded_at FROM datasets WHERE user_id=$1 ORDER BY uploaded_at DESC", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db query failed"})
		return
	}
	defer rows.Close()

	var datasets []map[string]interface{}
	for rows.Next() {
		var id int
		var filename, fileURL string
		var size int64
		var uploadedAt time.Time
		rows.Scan(&id, &filename, &fileURL, &size, &uploadedAt)

		datasets = append(datasets, map[string]interface{}{
			"id":         id,
			"filename":   filename,
			"file_url":   fileURL,
			"size_bytes": size,
			"uploaded_at": uploadedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"datasets": datasets})
}