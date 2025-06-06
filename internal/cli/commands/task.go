package commands

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/terzigolu/josepshbrain-go/internal/cli/interactive"
	"github.com/terzigolu/josepshbrain-go/pkg/models"
	"golang.org/x/term"
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
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List tasks for the active project",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			allProjects, _ := cmd.Flags().GetBool("all")
			status, _ := cmd.Flags().GetString("status")
			
			// Get active project unless --all flag is used
			var project models.Project
			if !allProjects {
				result := db.Where("is_active = ? AND deleted_at IS NULL", true).First(&project)
				if result.Error != nil {
					fmt.Println("❌ No active project found")
					fmt.Println("💡 Use 'jbraincli use <project>' to set an active project")
					fmt.Println("💡 Or use --all flag to see tasks from all projects")
					return
				}
			}

			// Build query for tasks
			query := db.Preload("Project")
			if !allProjects {
				query = query.Where("project_id = ?", project.ID)
			}
			if status != "" {
				query = query.Where("status = ?", strings.ToUpper(status))
			}

			var tasks []models.Task
			if err := query.Find(&tasks).Error; err != nil {
				log.Fatalf("Failed to fetch tasks: %v", err)
			}

			if len(tasks) == 0 {
				if !allProjects {
					fmt.Printf("📋 No tasks found in project '%s'\n", project.Name)
				} else {
					fmt.Println("📋 No tasks found in any project")
				}
				fmt.Println("💡 Create one with 'jbraincli task create <description>'")
				return
			}

			// Display beautiful task list
			displayTaskList(tasks, project.Name, allProjects, status)
		},
	}
	
	cmd.Flags().BoolP("all", "a", false, "Show tasks from all projects")
	cmd.Flags().StringP("status", "s", "", "Filter by status (TODO, IN_PROGRESS, IN_REVIEW, COMPLETED)")
	
	return cmd
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

// displayTaskList shows tasks in a beautiful, responsive format
func displayTaskList(tasks []models.Task, projectName string, allProjects bool, statusFilter string) {
	// Import terminal width detection
	var width int = 80 // default width
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width = w
	}

	// Header with project info
	if allProjects {
		if statusFilter != "" {
			fmt.Printf("📋 %s Tasks from All Projects (%d)\n", strings.ToUpper(statusFilter), len(tasks))
		} else {
			fmt.Printf("📋 All Tasks from All Projects (%d)\n", len(tasks))
		}
	} else {
		if statusFilter != "" {
			fmt.Printf("📋 %s Tasks - %s (%d)\n", strings.ToUpper(statusFilter), projectName, len(tasks))
		} else {
			fmt.Printf("📋 Tasks - %s (%d)\n", projectName, len(tasks))
		}
	}

	// Generate unique short IDs (reuse from kanban)
	uniqueIDs := generateUniqueShortIDsForTasks(tasks)

	// Responsive design
	if width < 100 {
		// Compact view for narrow terminals
		displayTaskListCompact(tasks, uniqueIDs, allProjects)
	} else {
		// Full table view for wide terminals
		displayTaskListTable(tasks, uniqueIDs, allProjects, width)
	}
}

// displayTaskListCompact shows tasks in compact format
func displayTaskListCompact(tasks []models.Task, uniqueIDs map[string]string, allProjects bool) {
	fmt.Println()
	for i, task := range tasks {
		// Priority and status icons
		priorityIcon := getPriorityIconForTask(task.Priority)
		statusIcon := getStatusIconForTask(task.Status)
		
		// Progress indicator
		progressBar := getProgressBar(task.Progress, 8)
		
		fmt.Printf("%s %s %s %s\n", 
			priorityIcon, 
			statusIcon, 
			uniqueIDs[task.ID.String()], 
			task.Description)
		
		if allProjects && task.Project != nil {
			fmt.Printf("   🏢 %s", task.Project.Name)
		}
		
		if task.Progress > 0 {
			fmt.Printf("   %s %d%%", progressBar, task.Progress)
		}
		
		fmt.Println()
		
		// Add separator between tasks (except last)
		if i < len(tasks)-1 {
			fmt.Println("   ─────────────────────────────────────")
		}
	}
}

// displayTaskListTable shows tasks in full table format  
func displayTaskListTable(tasks []models.Task, uniqueIDs map[string]string, allProjects bool, termWidth int) {
	// Calculate dynamic column widths
	idWidth := 12
	priorityWidth := 4
	statusWidth := 12
	progressWidth := 12
	projectWidth := 0
	if allProjects {
		projectWidth = 20
	}
	
	// Remaining width for description
	usedWidth := idWidth + priorityWidth + statusWidth + progressWidth + projectWidth + 8 // borders and spaces
	descWidth := termWidth - usedWidth
	if descWidth < 30 {
		descWidth = 30
	}

	// Table header
	fmt.Println()
	if allProjects {
		fmt.Printf("┌─%-*s─┬─%-*s─┬─%-*s─┬─%-*s─┬─%-*s─┬─%-*s─┐\n", 
			idWidth, strings.Repeat("─", idWidth),
			priorityWidth, strings.Repeat("─", priorityWidth),
			statusWidth, strings.Repeat("─", statusWidth),
			progressWidth, strings.Repeat("─", progressWidth),
			projectWidth, strings.Repeat("─", projectWidth),
			descWidth, strings.Repeat("─", descWidth))
		
		fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
			idWidth, "ID",
			priorityWidth, "PRI",
			statusWidth, "STATUS",
			progressWidth, "PROGRESS",
			projectWidth, "PROJECT",
			descWidth, "DESCRIPTION")
	} else {
		fmt.Printf("┌─%-*s─┬─%-*s─┬─%-*s─┬─%-*s─┬─%-*s─┐\n", 
			idWidth, strings.Repeat("─", idWidth),
			priorityWidth, strings.Repeat("─", priorityWidth),
			statusWidth, strings.Repeat("─", statusWidth),
			progressWidth, strings.Repeat("─", progressWidth),
			descWidth, strings.Repeat("─", descWidth))
		
		fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
			idWidth, "ID",
			priorityWidth, "PRI", 
			statusWidth, "STATUS",
			progressWidth, "PROGRESS",
			descWidth, "DESCRIPTION")
	}

	// Separator
	if allProjects {
		fmt.Printf("├─%-*s─┼─%-*s─┼─%-*s─┼─%-*s─┼─%-*s─┼─%-*s─┤\n",
			idWidth, strings.Repeat("─", idWidth),
			priorityWidth, strings.Repeat("─", priorityWidth),
			statusWidth, strings.Repeat("─", statusWidth),
			progressWidth, strings.Repeat("─", progressWidth),
			projectWidth, strings.Repeat("─", projectWidth),
			descWidth, strings.Repeat("─", descWidth))
	} else {
		fmt.Printf("├─%-*s─┼─%-*s─┼─%-*s─┼─%-*s─┼─%-*s─┤\n",
			idWidth, strings.Repeat("─", idWidth),
			priorityWidth, strings.Repeat("─", priorityWidth),
			statusWidth, strings.Repeat("─", statusWidth),
			progressWidth, strings.Repeat("─", progressWidth),
			descWidth, strings.Repeat("─", descWidth))
	}

	// Task rows
	for _, task := range tasks {
		priorityIcon := getPriorityIconForTask(task.Priority)
		statusIcon := getStatusIconForTask(task.Status)
		progressBar := getProgressBar(task.Progress, 10)
		
		shortID := uniqueIDs[task.ID.String()]
		description := truncateString(task.Description, descWidth)
		
		if allProjects {
			projectName := ""
			if task.Project != nil {
				projectName = truncateString(task.Project.Name, projectWidth)
			}
			
			fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
				idWidth, shortID,
				priorityWidth, priorityIcon,
				statusWidth, statusIcon,
				progressWidth, progressBar,
				projectWidth, projectName,
				descWidth, description)
		} else {
			fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │\n",
				idWidth, shortID,
				priorityWidth, priorityIcon,
				statusWidth, statusIcon,
				progressWidth, progressBar,
				descWidth, description)
		}
	}

	// Table footer
	if allProjects {
		fmt.Printf("└─%-*s─┴─%-*s─┴─%-*s─┴─%-*s─┴─%-*s─┴─%-*s─┘\n",
			idWidth, strings.Repeat("─", idWidth),
			priorityWidth, strings.Repeat("─", priorityWidth),
			statusWidth, strings.Repeat("─", statusWidth),
			progressWidth, strings.Repeat("─", progressWidth),
			projectWidth, strings.Repeat("─", projectWidth),
			descWidth, strings.Repeat("─", descWidth))
	} else {
		fmt.Printf("└─%-*s─┴─%-*s─┴─%-*s─┴─%-*s─┴─%-*s─┘\n",
			idWidth, strings.Repeat("─", idWidth),
			priorityWidth, strings.Repeat("─", priorityWidth),
			statusWidth, strings.Repeat("─", statusWidth),
			progressWidth, strings.Repeat("─", progressWidth),
			descWidth, strings.Repeat("─", descWidth))
	}
}

// Helper functions for task list display
func generateUniqueShortIDsForTasks(tasks []models.Task) map[string]string {
	uniqueIDs := make(map[string]string)
	usedShortIDs := make(map[string][]string)
	
	// First pass: try 8-character IDs
	for _, task := range tasks {
		fullID := task.ID.String()
		shortID := fullID[:8]
		usedShortIDs[shortID] = append(usedShortIDs[shortID], fullID)
	}
	
	// Second pass: resolve collisions
	for shortID, fullIDs := range usedShortIDs {
		if len(fullIDs) == 1 {
			uniqueIDs[fullIDs[0]] = shortID
		} else {
			for _, fullID := range fullIDs {
				uniqueLen := 8
				for uniqueLen < len(fullID) {
					candidate := fullID[:uniqueLen]
					isUnique := true
					for _, otherID := range fullIDs {
						if otherID != fullID && len(otherID) > uniqueLen && otherID[:uniqueLen] == candidate {
							isUnique = false
							break
						}
					}
					if isUnique {
						break
					}
					uniqueLen++
				}
				uniqueIDs[fullID] = fullID[:uniqueLen]
			}
		}
	}
	
	return uniqueIDs
}

func getPriorityIconForTask(priority string) string {
	icons := map[string]string{
		"H": "🔴",
		"M": "🟡",
		"L": "🟢",
	}
	if icon, exists := icons[priority]; exists {
		return icon
	}
	return "⚪"
}

func getStatusIconForTask(status string) string {
	icons := map[string]string{
		"TODO":        "📋",
		"IN_PROGRESS": "🚀", 
		"IN_REVIEW":   "👀",
		"COMPLETED":   "✅",
	}
	if icon, exists := icons[status]; exists {
		return icon
	}
	return "❓"
}

func getProgressBar(progress int, width int) string {
	if progress == 0 {
		return strings.Repeat("░", width)
	}
	if progress == 100 {
		return "✅ 100%"
	}
	
	filled := (progress * width) / 100
	bar := strings.Repeat("▓", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("%s %d%%", bar, progress)
} 