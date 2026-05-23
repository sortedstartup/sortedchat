//go:build ignore
// +build ignore

package main

import (
	"context"

	"github.com/sortedstartup/sortedagents"
)

func main() {
	ctx := context.Background()
	runner := sortedagents.NewRunner()
	agent := sortedagents.NewAgent("test", "test", "gpt-4o-mini", nil)
	session := sortedagents.NewSession()

	// Test new API signature
	message := sortedagents.Message{
		Role:    "user",
		Content: sortedagents.TextContent("test"),
	}
	_, _ = runner.Run(ctx, agent, message, 10, session)
	_ = runner.RunStream(ctx, agent, message, 10, session)
}
