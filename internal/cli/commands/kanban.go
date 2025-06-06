package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/terzigolu/josepshbrain-go/pkg/models"
	"golang.org/x/term"
	"gorm.io/gorm"
)

// NewKanbanCmd creates the kanban command
func NewKanbanCmd(db *gorm.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "kanban",
		Short: "Display tasks in a beautiful kanban board",
		Long:  "Show tasks organized by status in a full-width kanban board layout",
		Run: func(cmd *cobra.Command, args []string) {
			// Get active project
			var project models.Project
			result := db.Where("is_active = ? AND deleted_at IS NULL", true).First(&project)
			if result.Error != nil {
				fmt.Println("❌ No active project found")
				fmt.Println("💡 Use 'jbraincli use <project>' to set an active project")
				return
			}

			// Get all tasks for the active project
			var tasks []models.Task
			err := db.Where("project_id = ?", project.ID).Find(&tasks).Error
			if err != nil {
				fmt.Printf("❌ Error fetching tasks: %v\n", err)
				return
			}

			// Display kanban board
			displayKanbanBoard(tasks, project.Name)
		},
	}
}

func displayKanbanBoard(tasks []models.Task, projectName string) {
	// Get terminal width
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 120 // Default width
	}

	// Check if terminal is too narrow for kanban view
	if width < 80 {
		displayCompactTaskList(tasks, projectName)
		return
	}

	// Organize tasks by status
	statusColumns := map[string][]models.Task{
		"TODO":        {},
		"IN_PROGRESS": {},
		"IN_REVIEW":   {},
		"COMPLETED":   {},
	}

	for _, task := range tasks {
		if _, exists := statusColumns[task.Status]; exists {
			statusColumns[task.Status] = append(statusColumns[task.Status], task)
		}
	}

	// Calculate column width (4 columns + borders + padding)
	columnWidth := (width - 8) / 4 // 8 chars for borders and spacing
	
	// Ensure minimum column width
	if columnWidth < 20 {
		columnWidth = 20
	}

	// Header
	fmt.Printf("\n🎯 %s - Kanban Board\n\n", projectName)

	// Print top border
	printKanbanBorder(columnWidth, "top")

	// Print column headers with task counts
	fmt.Print("│")
	printCenteredText(fmt.Sprintf("📋 TODO (%d)", len(statusColumns["TODO"])), columnWidth)
	fmt.Print("│")
	printCenteredText(fmt.Sprintf("🚀 IN PROGRESS (%d)", len(statusColumns["IN_PROGRESS"])), columnWidth)
	fmt.Print("│")
	printCenteredText(fmt.Sprintf("👀 IN REVIEW (%d)", len(statusColumns["IN_REVIEW"])), columnWidth)
	fmt.Print("│")
	printCenteredText(fmt.Sprintf("✅ COMPLETED (%d)", len(statusColumns["COMPLETED"])), columnWidth)
	fmt.Println("│")

	// Print separator
	printKanbanBorder(columnWidth, "middle")

	// Find max tasks in any column for row count
	maxTasks := 0
	for _, tasks := range statusColumns {
		if len(tasks) > maxTasks {
			maxTasks = len(tasks)
		}
	}

	// Build unique short IDs for all tasks to avoid collisions
	allTasks := []models.Task{}
	for _, taskList := range statusColumns {
		allTasks = append(allTasks, taskList...)
	}
	uniqueIDs := generateUniqueShortIDs(allTasks)

	// Print task rows
	statuses := []string{"TODO", "IN_PROGRESS", "IN_REVIEW", "COMPLETED"}
	for i := 0; i < maxTasks; i++ {
		fmt.Print("│")
		for _, status := range statuses {
			tasks := statusColumns[status]
			if i < len(tasks) {
				taskText := formatTaskForKanbanWithID(tasks[i], uniqueIDs[tasks[i].ID.String()], columnWidth-2)
				fmt.Printf(" %-*s", columnWidth-2, taskText)
			} else {
				fmt.Printf(" %-*s", columnWidth-2, "")
			}
			fmt.Print(" │")
		}
		fmt.Println()
	}

	// Print bottom border
	printKanbanBorder(columnWidth, "bottom")

	// Print summary
	fmt.Printf("\n📊 Summary: %d TODO • %d IN PROGRESS • %d IN REVIEW • %d COMPLETED\n\n",
		len(statusColumns["TODO"]),
		len(statusColumns["IN_PROGRESS"]),
		len(statusColumns["IN_REVIEW"]),
		len(statusColumns["COMPLETED"]))
}

func printKanbanBorder(columnWidth int, position string) {
	var left, right, horizontal, junction string

	switch position {
	case "top":
		left, right, horizontal, junction = "┌", "┐", "─", "┬"
	case "middle":
		left, right, horizontal, junction = "├", "┤", "─", "┼"
	case "bottom":
		left, right, horizontal, junction = "└", "┘", "─", "┴"
	}

	fmt.Print(left)
	for i := 0; i < 4; i++ {
		fmt.Print(strings.Repeat(horizontal, columnWidth))
		if i < 3 {
			fmt.Print(junction)
		}
	}
	fmt.Println(right)
}

func printCenteredText(text string, width int) {
	textLen := len(text)
	if textLen >= width {
		fmt.Printf(" %-*s", width-2, truncateString(text, width-2))
		return
	}

	padding := (width - textLen) / 2
	fmt.Printf("%*s%s%*s", padding, "", text, width-textLen-padding, "")
}

func formatTaskForKanban(task models.Task, maxWidth int) string {
	// Priority indicator
	priorityIcon := map[string]string{
		"H": "🔴",
		"M": "🟡", 
		"L": "🟢",
	}

	icon := "⚪"
	if p, exists := priorityIcon[task.Priority]; exists {
		icon = p
	}

	// Smart ID truncation - start with 8 chars, extend if needed for uniqueness
	shortID := task.ID.String()[:8]
	
	// Progress indicator for non-TODO tasks
	progressIndicator := ""
	if task.Status != "TODO" && task.Progress > 0 {
		if task.Progress == 100 {
			progressIndicator = " ✅"
		} else if task.Progress >= 75 {
			progressIndicator = " ▓▓▓▓▓▓▓░"
		} else if task.Progress >= 50 {
			progressIndicator = " ▓▓▓▓▓░░░"
		} else if task.Progress >= 25 {
			progressIndicator = " ▓▓▓░░░░░"
		} else {
			progressIndicator = " ▓░░░░░░░"
		}
	}
	
	// Format: icon + short ID + description + progress
	prefix := fmt.Sprintf("%s %s ", icon, shortID)
	suffix := progressIndicator
	availableWidth := maxWidth - len(prefix) - len(suffix)
	
	if availableWidth <= 0 {
		return truncateString(prefix, maxWidth)
	}

	description := truncateString(task.Description, availableWidth)
	return prefix + description + suffix
}

// displayCompactTaskList shows a simple list view when terminal is too narrow
func displayCompactTaskList(tasks []models.Task, projectName string) {
	fmt.Printf("\n🎯 %s - Task List (Compact View)\n\n", projectName)
	
	statusOrder := []string{"TODO", "IN_PROGRESS", "IN_REVIEW", "COMPLETED"}
	statusIcons := map[string]string{
		"TODO":        "📋",
		"IN_PROGRESS": "🚀", 
		"IN_REVIEW":   "👀",
		"COMPLETED":   "✅",
	}
	
	for _, status := range statusOrder {
		statusTasks := []models.Task{}
		for _, task := range tasks {
			if task.Status == status {
				statusTasks = append(statusTasks, task)
			}
		}
		
		if len(statusTasks) > 0 {
			fmt.Printf("%s %s (%d)\n", statusIcons[status], status, len(statusTasks))
			for _, task := range statusTasks {
				priorityIcon := map[string]string{"H": "🔴", "M": "🟡", "L": "🟢"}[task.Priority]
				if priorityIcon == "" {
					priorityIcon = "⚪"
				}
				
				shortID := task.ID.String()[:8]
				description := truncateString(task.Description, 50)
				fmt.Printf("  %s %s %s\n", priorityIcon, shortID, description)
			}
			fmt.Println()
		}
	}
}

// generateUniqueShortIDs creates collision-free short IDs for a set of tasks
func generateUniqueShortIDs(tasks []models.Task) map[string]string {
	uniqueIDs := make(map[string]string)
	usedShortIDs := make(map[string][]string) // shortID -> list of full UUIDs
	
	// First pass: try 8-character IDs
	for _, task := range tasks {
		fullID := task.ID.String()
		shortID := fullID[:8]
		usedShortIDs[shortID] = append(usedShortIDs[shortID], fullID)
	}
	
	// Second pass: resolve collisions by extending length
	for shortID, fullIDs := range usedShortIDs {
		if len(fullIDs) == 1 {
			// No collision, use 8-char ID
			uniqueIDs[fullIDs[0]] = shortID
		} else {
			// Collision detected, extend until unique
			for _, fullID := range fullIDs {
				uniqueLen := 8
				for uniqueLen < len(fullID) {
					candidate := fullID[:uniqueLen]
					// Check if this length makes it unique among colliding IDs
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

// formatTaskForKanbanWithID formats a task with pre-calculated unique ID
func formatTaskForKanbanWithID(task models.Task, shortID string, maxWidth int) string {
	// Priority indicator
	priorityIcon := map[string]string{
		"H": "🔴",
		"M": "🟡", 
		"L": "🟢",
	}

	icon := "⚪"
	if p, exists := priorityIcon[task.Priority]; exists {
		icon = p
	}
	
	// Progress indicator for non-TODO tasks
	progressIndicator := ""
	if task.Status != "TODO" && task.Progress > 0 {
		if task.Progress == 100 {
			progressIndicator = " ✅"
		} else if task.Progress >= 75 {
			progressIndicator = " ▓▓▓▓▓▓▓░"
		} else if task.Progress >= 50 {
			progressIndicator = " ▓▓▓▓▓░░░"
		} else if task.Progress >= 25 {
			progressIndicator = " ▓▓▓░░░░░"
		} else {
			progressIndicator = " ▓░░░░░░░"
		}
	}
	
	// Format: icon + unique short ID + description + progress
	prefix := fmt.Sprintf("%s %s ", icon, shortID)
	suffix := progressIndicator
	availableWidth := maxWidth - len(prefix) - len(suffix)
	
	if availableWidth <= 0 {
		return truncateString(prefix, maxWidth)
	}

	description := truncateString(task.Description, availableWidth)
	return prefix + description + suffix
} 