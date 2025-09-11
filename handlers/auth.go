package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/edubank/db"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type SignupRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
	AccessCode string `json:"accessCode" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func SignupHandler(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"invalid input"})
		return
	}

	ctx := context.Background()

	// verify access code -> get university id
	var universityID int
	err := db.Pool.QueryRow(ctx,
		"SELECT id FROM universities WHERE code=$1",
		req.AccessCode).Scan(&universityID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error":"Invalid Access Code"})
		return
	}

	// hash password
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("bcrypt error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Server Error"})
		return
	}

	// insert user
	_, err = db.Pool.Exec(ctx,
		"INSERT INTO users (email, password_hash, university_id) VALUES ($1, $2, $3)",
		req.Email, string(hashBytes), universityID,
	)
	if err != nil {
		log.Printf("insert user error: %v", err)
		c.JSON(http.StatusConflict, gin.H{"error":"User already exists or DB error!"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": req.Email,
		"exp":   time.Now().Add(time.Hour * 24 * 7).Unix(), // expires in 7 days
	})

	tokenStr, _ := token.SignedString(jwtSecret)

	c.JSON(http.StatusOK, gin.H{
		"message": "Signup successful",
		"token":   tokenStr,
		"email":   req.Email,
	})
}

func LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"Invalid Input"})
		return
	}

	ctx := context.Background()
	var hash string
	err := db.Pool.QueryRow(ctx, "SELECT password_hash FROM users WHERE email=$1", req.Email).Scan(&hash)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error":"Invalid Credentials"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error":"Invalid Credentials"})
		return
	}

	// generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": req.Email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Token Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"email": req.Email,
	})
}
