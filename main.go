package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}


type OpenAIResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

type WebSocketMessage struct {
	Text string `json:"text"`
}

var (
	clients = make(map[*websocket.Conn]bool)
)

func main() {
	app := fiber.New()
	app.Static("/", "./static")
	app.Use(cors.New())
	
	app.Get("/", handleHome)
	app.Get("/ws", websocket.New(handleWebSocket))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server starting on :%s\n", port)
	app.Listen(":" + port)
}

func handleHome(c *fiber.Ctx) error {
	return c.SendFile("./static/index.html")
}

func handleWebSocket(c *websocket.Conn) {
	fmt.Println("New WebSocket connection established")
	
	clients[c] = true
	
	defer delete(clients, c)

	
	for {
		var msg WebSocketMessage
		
		err := c.ReadJSON(&msg)
		if err != nil {
			break
		}
		fmt.Printf("Received message: %+v\n", msg)
		go streamResponse(msg.Text, c)
	}
}

func streamResponse(message string, conn *websocket.Conn) {
    fmt.Printf("message from user:%s\n", message)
    ctx := context.Background()
    client, err := genai.NewClient(ctx, option.WithAPIKey("AIzaSyBHwRstvXpQd4ETadlep6b-JxRE1ncRjhw"))
    if err != nil {
        fmt.Println("Error creating genai client:", err)
        return
    }
    defer client.Close()

    model := client.GenerativeModel("gemini-1.5-flash")
    resp, err := model.GenerateContent(ctx, genai.Text(message))
    if err != nil {
        fmt.Println("Error generating content with genai:", err)
        return
    }

    for _, cand := range resp.Candidates {
        if cand.Content != nil {
            for _, part := range cand.Content.Parts {
                if conn.WriteJSON(WebSocketMessage{Text: partToString(part)}) != nil {
                    fmt.Println("Error sending message through WebSocket:", err)
                    return
                }else{
					fmt.Println("done")
				}
            }
        }
    }
}

func partToString(part genai.Part) string {
    switch p := part.(type) {
    case genai.Text:
        return string(p)
    case genai.Blob:
        
        return "Blob content (conversion needed)"
    default:
        fmt.Printf("Unsupported genai.Part type: %T\n", part)
        return ""
    }
}


