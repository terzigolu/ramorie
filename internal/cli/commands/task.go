package commands

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/terzigolu/josepshbrain-go/internal/cli/interactive"
	"github.com/terzigolu/josepshbrain-go/pkg/models"
	"gorm.io/gorm"
)

// NewTaskCmd creates the task command with all subcommands
func NewTaskCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Task management commands",
		Long:  "Create, list, update, and manage tasks",
	}

	// Add subcommands with database
	cmd.AddCommand(newTaskCreateCmd(db))
	cmd.AddCommand(newTaskListCmd(db))
	cmd.AddCommand(newTaskStartCmd(db))
	cmd.AddCommand(newTaskDoneCmd(db))
	cmd.AddCommand(newTaskInfoCmd(db))

	return cmd
}

// task create
func newTaskCreateCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create [description]",
		Short:   "Create a new task",
		Aliases: []string{"add"},
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool("interactive")
			
			var description string
			var priority string = "M" // default medium
			
			if isInteractive {
				// Interactive mode
				task, err := interactive.CreateTaskInteractive()
				if err != nil {
					log.Fatalf("Interactive task creation failed: %v", err)
				}
				description = task.Description
				priority = task.Priority
			} else {
				// Traditional CLI mode
				if len(args) == 0 {
					fmt.Println("❌ Description required (or use --interactive)")
					return
				}
				description = args[0]
			}
			
			// Get active project - require one to exist
			var project models.Project
			result := db.Where("is_active = ? AND deleted_at IS NULL", true).First(&project)
			if result.Error != nil {
				fmt.Println("❌ No active project found")
				fmt.Println("💡 Use 'jbraincli init <name>' to create a project first")
				return
			}

			// Create new task
			task := models.Task{
				ProjectID:   project.ID,
				Description: description,
				Status:      string(models.TaskStatusTODO),
				Priority:    priority,
				Progress:    0,
			}

			if err := db.Create(&task).Error; err != nil {
				log.Fatalf("Failed to create task: %v", err)
			}

			fmt.Printf("🔄 Created task: %s\n", description)
			fmt.Printf("✅ Task ID: %s\n", task.ID.String())
		},
	}
	
	// Add interactive flag
	cmd.Flags().BoolP("interactive", "i", false, "Use interactive mode for task creation")
	cmd.Args = cobra.MinimumNArgs(0) // Make args optional when using interactive
	
	return cmd
}

// task list
func newTaskListCmd(db *gorm.DB) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List all tasks",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			var tasks []models.Task
			if err := db.Preload("Project").Find(&tasks).Error; err != nil {
				log.Fatalf("Failed to fetch tasks: %v", err)
			}

			if len(tasks) == 0 {
				fmt.Println("📋 No tasks found. Create one with 'jbraincli task create <description>'")
				return
			}

			fmt.Println("📋 Task List:")
			fmt.Println("┌─────────────────────────────────────────┬───────────────────────────┬──────────────┬──────────┐")
			fmt.Println("│ ID                                      │ Description               │ Status       │ Priority │")
			fmt.Println("├─────────────────────────────────────────┼───────────────────────────┼──────────────┼──────────┤")
			
			for _, task := range tasks {
				fmt.Printf("│ %-39s │ %-25s │ %-12s │ %-8s │\n", 
					task.ID.String()[:8]+"...", 
					truncateString(task.Description, 25),
					string(task.Status),
					string(task.Priority))
			}
			fmt.Println("└─────────────────────────────────────────┴───────────────────────────┴──────────────┴──────────┘")
		},
	}
}

// task start
func newTaskStartCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [id]",
		Short: "Start working on a task",
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool("interactive")
			
			var task models.Task
			
			if isInteractive {
				// Interactive mode - select from TODO tasks
				var todoTasks []models.Task
				if err := db.Where("status = ?", "TODO").Find(&todoTasks).Error; err != nil {
					log.Fatalf("Failed to fetch TODO tasks: %v", err)
				}
				
				if len(todoTasks) == 0 {
					fmt.Println("📋 No TODO tasks available to start")
					return
				}
				
				selectedTask, err := interactive.SelectTask(todoTasks, "Select task to start:")
				if err != nil {
					log.Fatalf("Task selection failed: %v", err)
				}
				task = *selectedTask
			} else {
				// Traditional CLI mode
				if len(args) == 0 {
					fmt.Println("❌ Task ID required (or use --interactive)")
					return
				}
				taskID := args[0]
				
				if err := db.Where("id::text LIKE ?", taskID+"%").First(&task).Error; err != nil {
					log.Fatalf("Task not found: %v", err)
				}
			}

			task.Status = string(models.TaskStatusInProgress)
			if err := db.Save(&task).Error; err != nil {
				log.Fatalf("Failed to update task: %v", err)
			}

			fmt.Printf("▶️ Started task: %s\n", task.Description)
			fmt.Println("✅ Task status updated to IN_PROGRESS!")
		},
	}
	
	// Add interactive flag
	cmd.Flags().BoolP("interactive", "i", false, "Use interactive mode for task selection")
	
	return cmd
}

// task done
func newTaskDoneCmd(db *gorm.DB) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "done [id]",
		Short: "Mark task as completed",
		Args:  cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool("interactive")
			
			var task models.Task
			
			if isInteractive {
				// Interactive mode - select from active tasks
				var activeTasks []models.Task
				if err := db.Where("status IN (?)", []string{"IN_PROGRESS", "IN_REVIEW"}).Find(&activeTasks).Error; err != nil {
					log.Fatalf("Failed to fetch active tasks: %v", err)
				}
				
				if len(activeTasks) == 0 {
					fmt.Println("📋 No active tasks to complete")
					return
				}
				
				selectedTask, err := interactive.SelectTask(activeTasks, "Select task to complete:")
				if err != nil {
					log.Fatalf("Task selection failed: %v", err)
				}
				task = *selectedTask
			} else {
				// Traditional CLI mode
				if len(args) == 0 {
					fmt.Println("❌ Task ID required (or use --interactive)")
					return
				}
				taskID := args[0]
				
				if err := db.Where("id::text LIKE ?", taskID+"%").First(&task).Error; err != nil {
					log.Fatalf("Task not found: %v", err)
				}
			}

			task.Status = string(models.TaskStatusCompleted)
			task.Progress = 100
			if err := db.Save(&task).Error; err != nil {
				log.Fatalf("Failed to update task: %v", err)
			}

			fmt.Printf("✅ Task completed: %s\n", task.Description)
			fmt.Println("🎉 Great job!")
		},
	}
	
	// Add interactive flag
	cmd.Flags().BoolP("interactive", "i", false, "Use interactive mode for task selection")
	
	return cmd
}

// task info
func newTaskInfoCmd(db *gorm.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "info [id]",
		Short: "Show detailed task information",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			taskID := args[0]
			
			var task models.Task
			if err := db.Preload("Project").Preload("Annotations").Where("id::text LIKE ?", taskID+"%").First(&task).Error; err != nil {
				log.Fatalf("Task not found: %v", err)
			}

			fmt.Println("🔍 Task Details:")
			fmt.Println("================================================================================")
			fmt.Printf("📝 ID:          %s\n", task.ID.String())
			fmt.Printf("📋 Description: %s\n", task.Description)
			fmt.Printf("📊 Status:      %s\n", task.Status)
			fmt.Printf("⚡ Priority:    %s\n", task.Priority)
			fmt.Printf("📈 Progress:    %d%%\n", task.Progress)
			fmt.Printf("🏢 Project:     %s\n", task.Project.Name)
			fmt.Printf("📅 Created:     %s\n", task.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("🔄 Updated:     %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05"))
			
			if len(task.Annotations) > 0 {
				fmt.Printf("\n📝 Annotations (%d):\n", len(task.Annotations))
				for i, annotation := range task.Annotations {
					fmt.Printf("  %d. %s\n", i+1, annotation.Content)
					fmt.Printf("     📅 %s\n", annotation.CreatedAt.Format("2006-01-02 15:04:05"))
				}
			} else {
				fmt.Println("\n📝 Annotations: None")
			}
			
			fmt.Println("================================================================================")
		},
	}
}


 