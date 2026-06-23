// name: Joke History Workflow
// description: Generates two jokes sequentially via Agent A and Agent B, then combines them into a story via Agent C.
package main

import (
	"context"
	"don_consiglieri/pkg/primitives"
	"fmt"
	"log"
)

func main() {
	p := primitives.NewPipeline("Joke History Pipeline", func(ctx context.Context) error {
		fmt.Println("Starting Joke History Workflow...")

		// Step 1: Agent A tells a joke (max 5 lines)
		agentA := primitives.Agent[primitives.FoundationModelResponse]{
			Name:        "Agent A",
			Provider:    primitives.AgyProvider{},
			Description: "Joke Generator A",
			Model:       "gemini-3.1-pro",
			Prompt:      "Tell a funny joke. The joke must be at most 5 lines.",
		}

		fmt.Println("Running Agent A...")
		resA, err := agentA.Run(ctx)
		if err != nil {
			return fmt.Errorf("agent A failed: %w", err)
		}
		jokeA := resA.ModelResponse.Result()
		fmt.Printf("Agent A's Joke:\n%s\n\n", jokeA)

		// Step 2: Agent B tells another joke (max 5 lines)
		agentB := primitives.Agent[primitives.FoundationModelResponse]{
			Name:        "Agent B",
			Provider:    primitives.AgyProvider{},
			Description: "Joke Generator B",
			Model:       "gemini-3.1-pro",
			Prompt:      "Tell another funny joke. The joke must be at most 5 lines.",
		}

		fmt.Println("Running Agent B...")
		resB, err := agentB.Run(ctx)
		if err != nil {
			return fmt.Errorf("agent B failed: %w", err)
		}
		jokeB := resB.ModelResponse.Result()
		fmt.Printf("Agent B's Joke:\n%s\n\n", jokeB)

		// Step 3: Agent C creates a story using both jokes (max 20 lines)
		agentC := primitives.Agent[primitives.FoundationModelResponse]{
			Name:        "Agent C",
			Provider:    primitives.AgyProvider{},
			Description: "Story Writer",
			Model:       "gemini-3.1-pro",
			Prompt:      fmt.Sprintf("Write a short story (maximum 20 lines) that incorporates these two jokes:\n\nJoke 1:\n%s\n\nJoke 2:\n%s", jokeA, jokeB),
		}

		fmt.Println("Running Agent C...")
		resC, err := agentC.Run(ctx)
		if err != nil {
			return fmt.Errorf("agent C failed: %w", err)
		}
		story := resC.ModelResponse.Result()
		fmt.Printf("Agent C's Story:\n%s\n\n", story)

		fmt.Println("Joke History Workflow completed successfully!")
		return nil
	})

	// Create an orchestrator and add the pipeline
	orch := primitives.NewOrchestrator("Joke History Orchestrator")
	orch.AddWorkflow(p)

	// Run the orchestrator. It handles session initialization.
	if err := orch.Run(context.Background()); err != nil {
		log.Fatalf("Orchestrator failed: %v", err)
	}
}
