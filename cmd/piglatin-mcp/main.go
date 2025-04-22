package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"github.com/mocksi/temporal-mcp/internal/service/piglatin"
)

// PigLatinRequest defines the request structure for Pig Latin conversion
type PigLatinRequest struct {
	Sentence string `json:"sentence" jsonschema:"required,description=The sentence to convert"`
}

func main() {
	// Configure logger to write to stderr instead of stdout
	log.SetOutput(os.Stderr)
	log.Println("Starting Temporal MCP with Pig Latin example...")

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Create a new MCP server with stdio transport for AI model communication
	server := mcp.NewServer(stdio.NewStdioServerTransport())

	// Register the ToPigLatin tool
	err := server.RegisterTool("toPigLatin", "Convert a sentence to Pig Latin", func(args PigLatinRequest) (*mcp.ToolResponse, error) {
		if args.Sentence == "" {
			return mcp.NewToolResponse(mcp.NewTextContent("Please provide a sentence to convert.")), nil
		}

		pigLatinSentence := piglatin.ToPigLatin(args.Sentence)
		return mcp.NewToolResponse(mcp.NewTextContent(pigLatinSentence)), nil
	})
	if err != nil {
		log.Fatalf("Failed to register toPigLatin tool: %v", err)
	}

	// Register the FromPigLatin tool
	err = server.RegisterTool("fromPigLatin", "Convert a Pig Latin sentence back to English", func(args PigLatinRequest) (*mcp.ToolResponse, error) {
		if args.Sentence == "" {
			return mcp.NewToolResponse(mcp.NewTextContent("Please provide a Pig Latin sentence to convert.")), nil
		}

		englishSentence := piglatin.FromPigLatin(args.Sentence)
		return mcp.NewToolResponse(mcp.NewTextContent(englishSentence)), nil
	})
	if err != nil {
		log.Fatalf("Failed to register fromPigLatin tool: %v", err)
	}

	// Register a system prompt that explains how to use this MCP
	err = server.RegisterPrompt("system_prompt", "System prompt for the Pig Latin MCP", func(_ struct{}) (*mcp.PromptResponse, error) {
		systemPrompt := `You are now connected to a Temporal MCP (Model Context Protocol) server that provides Pig Latin conversion capabilities.

This MCP exposes the following tools:

1. toPigLatin - Converts English sentences to Pig Latin
   Usage: Call this tool with {"sentence": "Your English sentence here"}

2. fromPigLatin - Converts Pig Latin sentences back to English
   Usage: Call this tool with {"sentence": "Ouryay Igpay Atinlay entencesay erehay"}

Pig Latin Rules:
- For words that begin with consonants, all consonants before the first vowel are moved to the end of the word and "ay" is added
- For words that begin with vowels, "way" is added to the end of the word
- Capitalization and punctuation are preserved

Example:
"Hello world" → "Ellohay orldway"
"I am happy" → "Iway amway appyhay"

You can use these tools to help users convert text to and from Pig Latin.`

		return mcp.NewPromptResponse("system_prompt", mcp.NewPromptMessage(mcp.NewTextContent(systemPrompt), mcp.Role("system"))), nil
	})
	if err != nil {
		log.Fatalf("Failed to register system prompt: %v", err)
	}

	// Start the MCP server in a goroutine
	go func() {
		// Don't print to stdout as it interferes with JSON communication
		log.Printf("Pig Latin MCP server is running. Press Ctrl+C to stop.")
		if err := server.Serve(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for termination signal
	sig := <-sigCh
	log.Printf("Received signal %v, shutting down MCP server...", sig)
	log.Printf("Temporal MCP server has been stopped.")
}
