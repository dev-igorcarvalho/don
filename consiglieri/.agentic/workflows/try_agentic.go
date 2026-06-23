// name: Try Agentic Workflow
// description: An agentic workflow that runs an AI Agent using the AgyProvider to explain what a simple agent is in 50 lines.
package main

import (
	"context"
	"don_consiglieri/pkg/primitives"
	"fmt"
	"log"
)

func main() {
	agentA := primitives.Agent[primitives.FoundationModelResponse]{
		Name:        "agentA lindo -maravilhoso //voador",
		Provider:    primitives.AgyProvider{},
		Description: "Agentic test",
		Model:       "gemini-3.1-pro",
		Prompt:      "Explain what is simple Agent in 50 lines.",
	}
	// Create a simple pipeline that does something
	p := primitives.NewPipeline("Agentic Workflow", func(ctx context.Context) error {
		fmt.Println("Starting Agentic Workflow...")
		res, err := agentA.Run(ctx)
		if err != nil {
			return err
		}
		fmt.Println(res.ArtifactPath)
		fmt.Println("Agentic Workflow completed successfully!")
		return nil
	})

	// Create an orchestrator and add the pipeline
	orch := primitives.NewOrchestrator("Sample Orchestrator")
	orch.AddWorkflow(p)

	// Run the orchestrator. It handles session initialization.
	if err := orch.Run(context.Background()); err != nil {
		log.Fatalf("Orchestrator failed: %v", err)
	}
}
