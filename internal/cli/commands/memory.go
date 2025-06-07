package commands

import (
	"fmt"

	"github.com/terzigolu/josepshbrain-go/internal/service"
	"github.com/urfave/cli/v2"
)

func NewMemoryCommand() *cli.Command {
	return &cli.Command{
		Name:    "memory",
		Aliases: []string{"m"},
		Usage:   "Manage memories",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List all memories",
				Action: func(c *cli.Context) error {
					api, err := service.NewAPIService()
					if err != nil {
						fmt.Printf("Error: %v\n", err)
						return nil
					}
					memories, err := api.ListMemories()
					if err != nil {
						return err
					}
					// ... display memories
					for _, m := range memories {
						fmt.Printf("ID: %s, Content: %s\n", m.ID, m.Content)
					}
					return nil
				},
			},
			{
				Name:      "create",
				Usage:     "Create a new memory",
				ArgsUsage: "[content]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "project-id",
						Usage:    "Project ID for the memory",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						return fmt.Errorf("memory content is required")
					}
					content := c.Args().First()
					projectID := c.String("project-id")
					
					api, err := service.NewAPIService()
					if err != nil {
						fmt.Printf("Error: %v\n", err)
						return nil
					}
					memory, err := api.CreateMemory(content, projectID)
					if err != nil {
						return err
					}
					fmt.Printf("Memory created successfully: %s (ID: %s)\n", memory.Content, memory.ID)
					return nil
				},
			},
		},
	}
}