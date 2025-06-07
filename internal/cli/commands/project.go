package commands

import (
	"fmt"

	"github.com/terzigolu/josepshbrain-go/internal/service"
	"github.com/urfave/cli/v2"
)

func NewProjectCommand() *cli.Command {
	return &cli.Command{
		Name:    "project",
		Aliases: []string{"p"},
		Usage:   "Manage projects",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List all projects",
				Action: func(c *cli.Context) error {
					api, err := service.NewAPIService()
					if err != nil {
						fmt.Printf("Error: %v\n", err)
						return nil
					}
					projects, err := api.ListProjects()
					if err != nil {
						return err
					}
					// ... display projects
					for _, p := range projects {
						fmt.Printf("ID: %s, Name: %s\n", p.ID, p.Name)
					}
					return nil
				},
			},
			{
				Name:      "create",
				Usage:     "Create a new project",
				ArgsUsage: "[name]",
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						return fmt.Errorf("project name is required")
					}
					name := c.Args().First()
					api, err := service.NewAPIService()
					if err != nil {
						fmt.Printf("Error: %v\n", err)
						return nil
					}
					project, err := api.CreateProject(name)
					if err != nil {
						return err
					}
					fmt.Printf("Project created successfully: %s (ID: %s)\n", project.Name, project.ID)
					return nil
				},
			},
		},
	}
}


