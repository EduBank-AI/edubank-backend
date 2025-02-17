# edubank-BE

A backend service built using Go, designed to generate smart questions for tests for university faculty based on uploaded notes, lectures, and other course material.

## Installation

### Prerequisites

- Google Cloud API credentials

#### Clone the repository
```
git clone https://github.com/EduBank-AI/edubank-BE.git
cd eduBank-BE
```
#### Install dependencies
```
go mod tidy
```
#### Set up environment variables
```
touch .env
```

#### Edit .env file with your database, API keys, and secret configurations

```
GEMINI_API_KEY= YOUR_GEMINIAPI_KEY
GOOGLE_APPLICATION_CREDENTIALS= YOUR_GCP_CREDENTIALS_JSON
```

### Add a reference video to the assets folder in the main directory
- An example video.mp4 has been provided to get started

## Run the application

```
go run main.go
```
