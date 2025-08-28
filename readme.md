# Code Reviewer

A Go-based web service for automated code review using OpenRouter's AI API.

## Prerequisites

Before running the Code Reviewer service, ensure you have the following installed:

- **Go**: Version 1.16 or higher (Download from [golang.org](https://golang.org/dl/))
- **Git**: For cloning the repository (Download from [git-scm.com](https://git-scm.com/downloads))
- A valid **OpenRouter API Key** (Obtain from [openrouter.ai](https://openrouter.ai/))

### Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/krtu0p/code-reviewer.git
   cd code-reviewer
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Set up environment variables**:
   Create a `.env` file in the project root and add your OpenRouter API key:
   ```bash
   echo "OPENROUTER_API_KEY=your-api-key-here" > .env
   ```

## Usage

1. **Run the server**:
   ```bash
   go run main.go
   ```
   The server will start on `http://localhost:3000`.

2. **Send a code review request**:
   Use a tool like `curl` or Postman to send a POST request to the `/review` endpoint with a JSON body containing the code and language.

   Example using `curl`:
   ```bash
   curl -X POST http://localhost:3000/review \
   -H "Content-Type: application/json" \
   -d '{
       "language": "python",
       "code": "def hello():\n    print(\"Hello, World!\")"
   }'
   ```

3. **Response**:
   The API will return a JSON object with the code review results, including a summary, issues, suggestions, complexity, and a score (0-100).

   Example response:
   ```json
   {
       "summary": "The code is a simple Python function that prints a greeting.",
       "issues": [],
       "suggestions": ["Consider adding a docstring to describe the function."],
       "complexity": "simple",
       "score": 90
   }
   ```