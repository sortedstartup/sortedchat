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
	_, _ = runner.Run(ctx, agent, "test", 10, session)
	_ = runner.RunStream(ctx, agent, "test", 10, session)
}
