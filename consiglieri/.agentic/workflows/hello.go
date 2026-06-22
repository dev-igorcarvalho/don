// description: simple pipeline that does hello world
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"don_consiglieri/pkg/primitives"
)

func main() {
	// Create a simple pipeline that does something
	p := primitives.NewPipeline("Hello Workflow", func(ctx context.Context) error {
		fmt.Println("Starting Hello Workflow...")
		for i := 1; i <= 5; i++ {
			fmt.Printf("Step %d: Doing work...\n", i)
			time.Sleep(500 * time.Millisecond)
		}
		fmt.Println("Hello Workflow completed successfully!")
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
