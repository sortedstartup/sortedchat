package ai

import (
	"context"
	"fmt"
	"log"
	"os"
)

// ExampleUsage demonstrates how to use the AI abstraction
func ExampleUsage() {
	// Initialize model manager
	manager := NewModelManager()

	// Register providers (you'll need API keys in environment variables)
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		openaiProvider := NewOpenAIProvider(apiKey)
		manager.RegisterProvider(openaiProvider)
		log.Printf("Registered OpenAI provider")
	}

	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		claudeProvider := NewClaudeProvider(apiKey)
		manager.RegisterProvider(claudeProvider)
		log.Printf("Registered Claude provider")
	}

	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		geminiProvider := NewGeminiProvider(apiKey)
		manager.RegisterProvider(geminiProvider)
		log.Printf("Registered Gemini provider")
	}

	// Example 1: Simple text chat with OpenAI
	if provider, exists := manager.GetProvider("openai"); exists {
		fmt.Println("=== Example 1: Simple text chat with OpenAI ===")

		messages := []ChatMessage{
			NewTextMessage("system", "You are a helpful assistant"),
			NewTextMessage("user", "What is the capital of France?"),
		}

		request := ChatRequest{
			Model:    "gpt-4o-mini",
			Messages: messages,
			Stream:   true,
		}

		stream, err := provider.Chat(context.Background(), request)
		if err != nil {
			log.Printf("Error: %v", err)
		} else {
			for response := range stream {
				if response.Type == "text_delta" {
					fmt.Print(response.Delta)
				} else if response.Type == "completion" {
					fmt.Printf("\n[Completed - Input: %d, Output: %d tokens]\n",
						response.InputTokens, response.OutputTokens)
				}
			}
		}
	}

	// Example 2: Multimodal chat with image
	if provider, exists := manager.GetProvider("openai"); exists {
		fmt.Println("\n=== Example 2: Multimodal chat with image ===")

		// Create a multimodal message with text and image
		contents := []MessageContent{
			NewTextContent("What do you see in this image?"),
			NewImageContent("data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEAAA..."), // truncated base64
		}

		messages := []ChatMessage{
			NewTextMessage("system", "You are a helpful assistant that can analyze images"),
			NewMultimodalMessage("user", contents),
		}

		request := ChatRequest{
			Model:    "gpt-4o",
			Messages: messages,
			Stream:   true,
		}

		stream, err := provider.Chat(context.Background(), request)
		if err != nil {
			log.Printf("Error: %v", err)
		} else {
			for response := range stream {
				if response.Type == "text_delta" {
					fmt.Print(response.Delta)
				} else if response.Type == "completion" {
					fmt.Printf("\n[Completed]\n")
				}
			}
		}
	}

	// Example 3: Using text with embedded image references
	fmt.Println("\n=== Example 3: Text with embedded image references ===")

	text := `Please analyze this image: [image:base64:iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==] 
	and tell me what you see. Also, what do you think about this other image: [image:url:https://example.com/image.jpg]?`

	contents, err := ParseMultimodalInput(text)
	if err != nil {
		log.Printf("Error parsing multimodal input: %v", err)
	} else {
		fmt.Printf("Parsed %d content pieces:\n", len(contents))
		for i, content := range contents {
			fmt.Printf("  %d. Type: %s\n", i+1, content.Type)
			if content.Type == ContentTypeText {
				fmt.Printf("     Text: %s\n", content.Text[:min(50, len(content.Text))])
			} else {
				fmt.Printf("     Image: %s\n", content.ImageURL[:min(50, len(content.ImageURL))])
			}
		}
	}

	// Example 4: List available models from all providers
	fmt.Println("\n=== Example 4: Available models ===")
	supportedModels := manager.GetSupportedModels()
	for providerName, models := range supportedModels {
		fmt.Printf("%s models: %v\n", providerName, models)
	}
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ExampleWithClaude shows how to use Claude specifically
func ExampleWithClaude() {
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		provider := NewClaudeProvider(apiKey)

		messages := []ChatMessage{
			NewTextMessage("user", "Explain quantum computing in simple terms"),
		}

		request := ChatRequest{
			Model:       "claude-3-5-sonnet-20241022",
			Messages:    messages,
			Stream:      true,
			Temperature: 0.7,
			MaxTokens:   1000,
		}

		stream, err := provider.Chat(context.Background(), request)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}

		fmt.Println("=== Claude Response ===")
		for response := range stream {
			if response.Type == "text_delta" {
				fmt.Print(response.Delta)
			} else if response.Type == "completion" {
				fmt.Printf("\n[Completed]\n")
			} else if response.Type == "error" {
				fmt.Printf("Error: %s\n", response.Error)
			}
		}
	}
}

// ExampleWithGemini shows how to use Gemini specifically
func ExampleWithGemini() {
	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		provider := NewGeminiProvider(apiKey)

		messages := []ChatMessage{
			NewTextMessage("user", "Write a short poem about artificial intelligence"),
		}

		request := ChatRequest{
			Model:       "gemini-1.5-flash",
			Messages:    messages,
			Stream:      true,
			Temperature: 0.8,
		}

		stream, err := provider.Chat(context.Background(), request)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}

		fmt.Println("=== Gemini Response ===")
		for response := range stream {
			if response.Type == "text_delta" {
				fmt.Print(response.Delta)
			} else if response.Type == "completion" {
				fmt.Printf("\n[Completed]\n")
			} else if response.Type == "error" {
				fmt.Printf("Error: %s\n", response.Error)
			}
		}
	}
}
