# EduBank Backend

EduBank Backend is the core service powering the **EduBank platform** — an AI-driven educational tool designed for university faculty to automatically generate smart questions from uploaded lectures, notes, and other course material.

Built using **Go (Golang)**, it leverages powerful AI models (such as Google's Gemini) and video/audio processing to streamline test creation and content understanding.



## ✨ Features

- 📚 Lecture & Note Processing – Accepts PDF, DOCX, and video formats.
- 🧠 AI-Powered Question Generation – Uses Gemini API to generate context-aware questions.
- 🎞️ Video to Text Conversion – Transcribes lecture videos for use in QnA and question generation.
- 🔐 Environment-based Config – Securely handles API keys and credentials via `.env`.



## 🛠️ Tech Stack

- **Language**: Go (Golang)
- **AI Integration**: Gemini API (Google Cloud)
- **Transcription**: Google Cloud Speech-to-Text
- **Environment** Handling: godotenv
- **Build Tool**: Make (optional)



## 🚀 Getting Started

### Prerequisites

- Golang (v1.20 or later)
- Google Cloud Account (GCP)
- Gemini API Key

### Installation

#### 1. Clone the Repository

```bash
git clone https://github.com/EduBank-AI/edubank-backend.git
cd edubank-backend
```

#### 2. Install dependencies

```bash
go mod tidy
```

#### 3. Set Up Environment Variables

Create a `.env` file in the root directory:
```bash
GEMINI_API_KEY=your_gemini_api_key
GOOGLE_APPLICATION_CREDENTIALS=your_gcp_credentials.json
PORT=8080
```

> [!NOTE]
> The `GOOGLE_APPLICATION_CREDENTIALS` should point to your downloaded GCP JSON credentials file..



## 📦 Scripts

| Command | Description |
| --- | --- |
| `go run main.go` | Run using GO |
| `make build` | Build using make for production |



## 🙌 Contributing

Contributions are welcome! Feel free to fork the repository and submit a pull request.

Steps:
1. Fork the project
2. Create your feature branch `git checkout -b feature/my-feature`
3. Commit your changes `git commit -m 'Add some feature'`
4. Push to the branch `git push origin feature/my-feature`
5. Open a pull request


## 📄 License
This project is licensed under the MIT License.



## 📬 Contact
For issues, feature requests, or feedback, please open a GitHub issue.
