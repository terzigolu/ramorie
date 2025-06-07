package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/terzigolu/josepshbrain-go/internal/service"
	"github.com/urfave/cli/v2"
)

func NewTaskCommand() *cli.Command {
	return &cli.Command{
		Name:    "task",
		Aliases: []string{"t"},
		Usage:   "Manage tasks",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List all tasks",
				Action: func(c *cli.Context) error {
					api, err := service.NewAPIService()
					if err != nil {
						fmt.Printf("Error: %v\n", err)
						return nil
					}
					// For now, we list tasks for all projects.
					// A mechanism to set an active project will be needed.
					tasks, err := api.ListTasks("")
					if err != nil {
						return err
					}

					if len(tasks) == 0 {
						fmt.Println("No tasks found.")
						return nil
					}

					w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
					fmt.Fprintln(w, "ID\tTITLE\tSTATUS\tPROJECT")
					for _, task := range tasks {
						projectInfo := "N/A"
						if task.Project != nil {
							projectInfo = task.Project.Name
						}
						fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
							task.ID.String()[:8],
							task.Title,
							task.Status,
							projectInfo)
					}
					w.Flush()
					return nil
				},
			},
			{
				Name:      "create",
				Usage:     "Create a new task",
				ArgsUsage: "[title]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "project-id",
						Aliases:  []string{"p"},
						Usage:    "Project ID for the task",
						Required: true,
					},
					&cli.StringFlag{
						Name:    "description",
						Aliases: []string{"d"},
						Usage:   "Task description",
						Value:   "",
					},
					&cli.StringFlag{
						Name:    "priority",
						Usage:   "Task priority (L, M, H, C)",
						Value:   "L",
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						return fmt.Errorf("task title is required")
					}
					title := c.Args().First()
					projectID := c.String("project-id")
					description := c.String("description")
					priority := c.String("priority")
					
					api, err := service.NewAPIService()
					if err != nil {
						fmt.Printf("Error: %v\n", err)
						return nil
					}
					
					task, err := api.CreateTask(projectID, title, description, priority, nil)
					if err != nil {
						return err
					}
					fmt.Printf("Task created successfully: %s (ID: %s)\n", task.Title, task.ID.String()[:8])
					return nil
				},
			},
		},
	}
}