package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/terzigolu/josepshbrain-go/internal/api"
	"github.com/terzigolu/josepshbrain-go/internal/config"
)

type toolDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

func ToolDefinitions() []toolDef {
	return []toolDef{
		{
			Name:        "create_task",
			Description: "Yeni bir görev oluştur",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"description": map[string]interface{}{"type": "string"}, "priority": map[string]interface{}{"type": "string"}, "project": map[string]interface{}{"type": "string"}}, "required": []string{"description"}},
		},
		{
			Name:        "list_tasks",
			Description: "Görevleri listele",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"status": map[string]interface{}{"type": "string"}, "project": map[string]interface{}{"type": "string"}, "limit": map[string]interface{}{"type": "number"}}},
		},
		{
			Name:        "search_tasks",
			Description: "Görevlerde keyword arama yap",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"query": map[string]interface{}{"type": "string"}, "status": map[string]interface{}{"type": "string"}, "project": map[string]interface{}{"type": "string"}, "tag": map[string]interface{}{"type": "string"}, "limit": map[string]interface{}{"type": "number"}}, "required": []string{"query"}},
		},
		{
			Name:        "get_next_tasks",
			Description: "Sıradaki görevleri öncelik sırasına göre getir",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"count": map[string]interface{}{"type": "number"}, "project": map[string]interface{}{"type": "string"}, "tag": map[string]interface{}{"type": "string"}}},
		},
		{
			Name:        "get_task",
			Description: "Görev detaylarını getir",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"taskId": map[string]interface{}{"type": "string"}}, "required": []string{"taskId"}},
		},
		{
			Name:        "start_task",
			Description: "Görevi başlat (IN_PROGRESS)",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"taskId": map[string]interface{}{"type": "string"}}, "required": []string{"taskId"}},
		},
		{
			Name:        "complete_task",
			Description: "Görevi tamamla (COMPLETED)",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"taskId": map[string]interface{}{"type": "string"}}, "required": []string{"taskId"}},
		},
		{
			Name:        "update_task_status",
			Description: "Görev durumunu güncelle",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"taskId": map[string]interface{}{"type": "string"}, "status": map[string]interface{}{"type": "string"}}, "required": []string{"taskId", "status"}},
		},
		{
			Name:        "update_progress",
			Description: "Görev ilerleme durumunu güncelle (0-100)",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"taskId": map[string]interface{}{"type": "string"}, "progress": map[string]interface{}{"type": "number"}}, "required": []string{"taskId", "progress"}},
		},
		{
			Name:        "delete_task",
			Description: "Görevi sil",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"taskId": map[string]interface{}{"type": "string"}}, "required": []string{"taskId"}},
		},
		{
			Name:        "add_task_note",
			Description: "Göreve not ekle (annotation)",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"taskId": map[string]interface{}{"type": "string"}, "note": map[string]interface{}{"type": "string"}}, "required": []string{"taskId", "note"}},
		},
		{
			Name:        "create_subtask",
			Description: "Bir göreve alt görev ekle",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"parentTaskId": map[string]interface{}{"type": "string"}, "description": map[string]interface{}{"type": "string"}}, "required": []string{"parentTaskId", "description"}},
		},
		{
			Name:        "bulk_start_tasks",
			Description: "Birden fazla görevi tek seferde başlat",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"taskIds": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}}}, "required": []string{"taskIds"}},
		},
		{
			Name:        "bulk_complete_tasks",
			Description: "Birden fazla görevi tek seferde tamamla",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"taskIds": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}}}, "required": []string{"taskIds"}},
		},
		{
			Name:        "bulk_delete_tasks",
			Description: "Birden fazla görevi tek seferde sil",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"taskIds": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}}}, "required": []string{"taskIds"}},
		},
		{
			Name:        "list_projects",
			Description: "Projeleri listele",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
		{
			Name:        "create_project",
			Description: "Yeni proje oluştur",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"name": map[string]interface{}{"type": "string"}, "description": map[string]interface{}{"type": "string"}}, "required": []string{"name"}},
		},
		{
			Name:        "set_active_project",
			Description: "Aktif projeyi değiştir",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"projectName": map[string]interface{}{"type": "string"}}, "required": []string{"projectName"}},
		},
		{
			Name:        "add_memory",
			Description: "Yeni bir hafıza/not ekle",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"content": map[string]interface{}{"type": "string"}, "project": map[string]interface{}{"type": "string"}}, "required": []string{"content"}},
		},
		{
			Name:        "list_memories",
			Description: "Hafızaları listele",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"project": map[string]interface{}{"type": "string"}, "term": map[string]interface{}{"type": "string"}, "limit": map[string]interface{}{"type": "number"}}},
		},
		{
			Name:        "get_task_memories",
			Description: "Bir görev ile ilişkili hafıza öğelerini getir",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"taskId": map[string]interface{}{"type": "string"}}, "required": []string{"taskId"}},
		},
		{
			Name:        "memory_tasks",
			Description: "Bir hafıza ile ilişkili görevleri getir",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"memoryId": map[string]interface{}{"type": "string"}}, "required": []string{"memoryId"}},
		},
		{
			Name:        "create_memory_task_link",
			Description: "Görev-hafıza linki oluştur (manual)",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"taskId": map[string]interface{}{"type": "string"}, "memoryId": map[string]interface{}{"type": "string"}, "relationType": map[string]interface{}{"type": "string"}}, "required": []string{"taskId", "memoryId"}},
		},
		{
			Name:        "get_memory",
			Description: "Hafıza detaylarını getir",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"memoryId": map[string]interface{}{"type": "string"}}, "required": []string{"memoryId"}},
		},
		{
			Name:        "get_stats",
			Description: "Görev istatistiklerini getir",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"project": map[string]interface{}{"type": "string"}}},
		},
		{
			Name:        "get_history",
			Description: "Son X günün görev aktivitesini getir",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"days": map[string]interface{}{"type": "number"}, "project": map[string]interface{}{"type": "string"}}},
		},
		{
			Name:        "analyze_task_risks",
			Description: "Görev için risk analizi yap",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"taskId": map[string]interface{}{"type": "string"}}, "required": []string{"taskId"}},
		},
		{
			Name:        "analyze_task_dependencies",
			Description: "Görev için bağımlılık analizi yap",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{"taskId": map[string]interface{}{"type": "string"}}, "required": []string{"taskId"}},
		},
	}
}

func CallTool(client *api.Client, name string, args map[string]interface{}) (interface{}, error) {
	switch name {
	case "create_task":
		description, _ := args["description"].(string)
		description = strings.TrimSpace(description)
		if description == "" {
			return nil, errors.New("description is required")
		}
		priority, _ := args["priority"].(string)
		priority = normalizePriority(priority)
		projectIdentifier, _ := args["project"].(string)
		projectID, err := resolveProjectID(client, projectIdentifier)
		if err != nil {
			return nil, err
		}
		task, err := client.CreateTask(projectID, description, "", priority)
		if err != nil {
			return nil, err
		}
		return task, nil

	case "list_tasks":
		status, _ := args["status"].(string)
		projectIdentifier, _ := args["project"].(string)
		projectID := ""
		if strings.TrimSpace(projectIdentifier) != "" {
			pid, err := resolveProjectID(client, projectIdentifier)
			if err != nil {
				return nil, err
			}
			projectID = pid
		}
		tasks, err := client.ListTasks(projectID, strings.TrimSpace(status))
		if err != nil {
			return nil, err
		}
		limit := toInt(args["limit"])
		if limit > 0 && limit < len(tasks) {
			tasks = tasks[:limit]
		}
		return tasks, nil

	case "search_tasks":
		query, _ := args["query"].(string)
		query = strings.TrimSpace(query)
		if query == "" {
			return nil, errors.New("query is required")
		}
		status, _ := args["status"].(string)
		projectIdentifier, _ := args["project"].(string)
		tag, _ := args["tag"].(string)
		limit := toInt(args["limit"])

		projectID := ""
		if strings.TrimSpace(projectIdentifier) != "" {
			pid, err := resolveProjectID(client, projectIdentifier)
			if err != nil {
				return nil, err
			}
			projectID = pid
		}

		tags := []string{}
		if strings.TrimSpace(tag) != "" {
			tags = append(tags, strings.TrimSpace(tag))
		}

		tasks, err := client.ListTasksQuery(projectID, strings.TrimSpace(status), query, nil, tags)
		if err != nil {
			return nil, err
		}
		if limit > 0 && limit < len(tasks) {
			tasks = tasks[:limit]
		}
		return tasks, nil

	case "get_next_tasks":
		count := toInt(args["count"])
		if count <= 0 {
			count = 5
		}
		projectIdentifier, _ := args["project"].(string)
		tag, _ := args["tag"].(string)

		projectID := ""
		if strings.TrimSpace(projectIdentifier) != "" {
			pid, err := resolveProjectID(client, projectIdentifier)
			if err != nil {
				return nil, err
			}
			projectID = pid
		}

		tags := []string{}
		if strings.TrimSpace(tag) != "" {
			tags = append(tags, strings.TrimSpace(tag))
		}

		// Default to TODO tasks
		tasks, err := client.ListTasksQuery(projectID, "TODO", "", nil, tags)
		if err != nil {
			return nil, err
		}

		sort.Slice(tasks, func(i, j int) bool {
			pi := priorityRank(tasks[i].Priority)
			pj := priorityRank(tasks[j].Priority)
			if pi != pj {
				return pi > pj
			}
			return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
		})

		if count < len(tasks) {
			tasks = tasks[:count]
		}
		return tasks, nil

	case "get_task":
		taskID, _ := args["taskId"].(string)
		taskID = strings.TrimSpace(taskID)
		if taskID == "" {
			return nil, errors.New("taskId is required")
		}
		return client.GetTask(taskID)

	case "start_task":
		taskID, _ := args["taskId"].(string)
		taskID = strings.TrimSpace(taskID)
		if taskID == "" {
			return nil, errors.New("taskId is required")
		}
		if err := client.StartTask(taskID); err != nil {
			return nil, err
		}
		return map[string]interface{}{"ok": true}, nil

	case "complete_task":
		taskID, _ := args["taskId"].(string)
		taskID = strings.TrimSpace(taskID)
		if taskID == "" {
			return nil, errors.New("taskId is required")
		}
		if err := client.CompleteTask(taskID); err != nil {
			return nil, err
		}
		return map[string]interface{}{"ok": true}, nil

	case "update_task_status":
		taskID, _ := args["taskId"].(string)
		status, _ := args["status"].(string)
		taskID = strings.TrimSpace(taskID)
		status = strings.TrimSpace(status)
		if taskID == "" || status == "" {
			return nil, errors.New("taskId and status are required")
		}
		return client.UpdateTask(taskID, map[string]interface{}{"status": status})

	case "update_progress":
		taskID, _ := args["taskId"].(string)
		taskID = strings.TrimSpace(taskID)
		progress := toInt(args["progress"])
		if taskID == "" {
			return nil, errors.New("taskId is required")
		}
		if progress < 0 || progress > 100 {
			return nil, errors.New("progress must be between 0 and 100")
		}
		return client.UpdateTask(taskID, map[string]interface{}{"progress": progress})

	case "delete_task":
		taskID, _ := args["taskId"].(string)
		taskID = strings.TrimSpace(taskID)
		if taskID == "" {
			return nil, errors.New("taskId is required")
		}
		if err := client.DeleteTask(taskID); err != nil {
			return nil, err
		}
		return map[string]interface{}{"ok": true}, nil

	case "add_task_note":
		taskID, _ := args["taskId"].(string)
		note, _ := args["note"].(string)
		taskID = strings.TrimSpace(taskID)
		note = strings.TrimSpace(note)
		if taskID == "" || note == "" {
			return nil, errors.New("taskId and note are required")
		}
		return client.CreateAnnotation(taskID, note)

	case "create_subtask":
		parentTaskID, _ := args["parentTaskId"].(string)
		description, _ := args["description"].(string)
		parentTaskID = strings.TrimSpace(parentTaskID)
		description = strings.TrimSpace(description)
		if parentTaskID == "" || description == "" {
			return nil, errors.New("parentTaskId and description are required")
		}
		return client.CreateSubtask(parentTaskID, description)

	case "bulk_start_tasks":
		ids, err := resolveTaskIDList(client, args["taskIds"])
		if err != nil {
			return nil, err
		}
		status := "IN_PROGRESS"
		if err := client.BulkUpdateTasks(ids, &status, nil, nil); err != nil {
			return nil, err
		}
		return map[string]interface{}{"ok": true, "count": len(ids)}, nil

	case "bulk_complete_tasks":
		ids, err := resolveTaskIDList(client, args["taskIds"])
		if err != nil {
			return nil, err
		}
		status := "COMPLETED"
		if err := client.BulkUpdateTasks(ids, &status, nil, nil); err != nil {
			return nil, err
		}
		return map[string]interface{}{"ok": true, "count": len(ids)}, nil

	case "bulk_delete_tasks":
		ids, err := resolveTaskIDList(client, args["taskIds"])
		if err != nil {
			return nil, err
		}
		if err := client.BulkDeleteTasks(ids); err != nil {
			return nil, err
		}
		return map[string]interface{}{"ok": true, "count": len(ids)}, nil

	case "list_projects":
		return client.ListProjects()

	case "create_project":
		name, _ := args["name"].(string)
		description, _ := args["description"].(string)
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, errors.New("name is required")
		}
		return client.CreateProject(name, strings.TrimSpace(description))

	case "set_active_project":
		projectName, _ := args["projectName"].(string)
		projectName = strings.TrimSpace(projectName)
		if projectName == "" {
			return nil, errors.New("projectName is required")
		}
		projects, err := client.ListProjects()
		if err != nil {
			return nil, err
		}
		for _, p := range projects {
			if p.Name == projectName || strings.HasPrefix(p.ID.String(), projectName) {
				if err := client.SetProjectActive(p.ID.String()); err != nil {
					return nil, err
				}
				cfg, _ := config.LoadConfig()
				if cfg == nil {
					cfg = &config.Config{}
				}
				cfg.ActiveProjectID = p.ID.String()
				_ = config.SaveConfig(cfg)
				return map[string]interface{}{"ok": true, "project_id": p.ID.String(), "name": p.Name}, nil
			}
		}
		return nil, errors.New("project not found")

	case "add_memory":
		content, _ := args["content"].(string)
		content = strings.TrimSpace(content)
		if content == "" {
			return nil, errors.New("content is required")
		}
		projectIdentifier, _ := args["project"].(string)
		projectID, err := resolveProjectID(client, projectIdentifier)
		if err != nil {
			return nil, err
		}
		return client.CreateMemory(projectID, content)

	case "list_memories":
		projectIdentifier, _ := args["project"].(string)
		term, _ := args["term"].(string)
		projectID := ""
		if strings.TrimSpace(projectIdentifier) != "" {
			pid, err := resolveProjectID(client, projectIdentifier)
			if err != nil {
				return nil, err
			}
			projectID = pid
		}
		memories, err := client.ListMemories(projectID, "")
		if err != nil {
			return nil, err
		}
		term = strings.TrimSpace(term)
		if term != "" {
			filtered := memories[:0]
			for _, m := range memories {
				if strings.Contains(strings.ToLower(m.Content), strings.ToLower(term)) {
					filtered = append(filtered, m)
				}
			}
			memories = filtered
		}
		limit := toInt(args["limit"])
		if limit > 0 && limit < len(memories) {
			memories = memories[:limit]
		}
		return memories, nil

	case "get_memory":
		memoryID, _ := args["memoryId"].(string)
		memoryID = strings.TrimSpace(memoryID)
		if memoryID == "" {
			return nil, errors.New("memoryId is required")
		}
		return client.GetMemory(memoryID)

	case "get_task_memories":
		taskID, _ := args["taskId"].(string)
		taskID = strings.TrimSpace(taskID)
		if taskID == "" {
			return nil, errors.New("taskId is required")
		}
		return client.ListTaskMemories(taskID)

	case "memory_tasks":
		memoryID, _ := args["memoryId"].(string)
		memoryID = strings.TrimSpace(memoryID)
		if memoryID == "" {
			return nil, errors.New("memoryId is required")
		}
		return client.ListMemoryTasks(memoryID)

	case "create_memory_task_link":
		taskID, _ := args["taskId"].(string)
		memoryID, _ := args["memoryId"].(string)
		relationType, _ := args["relationType"].(string)
		taskID = strings.TrimSpace(taskID)
		memoryID = strings.TrimSpace(memoryID)
		relationType = strings.TrimSpace(relationType)
		if taskID == "" || memoryID == "" {
			return nil, errors.New("taskId and memoryId are required")
		}
		b, err := client.CreateMemoryTaskLink(taskID, memoryID, relationType)
		if err != nil {
			return nil, err
		}
		var out interface{}
		if err := json.Unmarshal(b, &out); err != nil {
			return map[string]interface{}{"ok": true}, nil
		}
		return out, nil

	case "get_stats":
		b, err := client.Request("GET", "/reports/stats", nil)
		if err != nil {
			return nil, err
		}
		var out interface{}
		if err := json.Unmarshal(b, &out); err != nil {
			return nil, fmt.Errorf("invalid stats response")
		}
		return out, nil

	case "get_history":
		days := toInt(args["days"])
		if days == 0 {
			days = 7
		}
		endpoint := "/reports/history"
		if days > 0 {
			endpoint = fmt.Sprintf("/reports/history?days=%d", days)
		}
		b, err := client.Request("GET", endpoint, nil)
		if err != nil {
			return nil, err
		}
		var out interface{}
		if err := json.Unmarshal(b, &out); err != nil {
			return nil, fmt.Errorf("invalid history response")
		}
		return out, nil

	case "analyze_task_risks":
		taskID, _ := args["taskId"].(string)
		taskID = strings.TrimSpace(taskID)
		if taskID == "" {
			return nil, errors.New("taskId is required")
		}
		return client.AIRisks(taskID)

	case "analyze_task_dependencies":
		taskID, _ := args["taskId"].(string)
		taskID = strings.TrimSpace(taskID)
		if taskID == "" {
			return nil, errors.New("taskId is required")
		}
		return client.AIDependencies(taskID)

	default:
		return nil, errors.New("tool not implemented")
	}
}

func priorityRank(p string) int {
	switch strings.ToUpper(strings.TrimSpace(p)) {
	case "H", "HIGH":
		return 3
	case "M", "MEDIUM":
		return 2
	case "L", "LOW":
		return 1
	default:
		return 0
	}
}

func resolveTaskIDList(client *api.Client, v interface{}) ([]string, error) {
	arr, ok := v.([]interface{})
	if !ok {
		return nil, errors.New("taskIds must be an array")
	}
	if len(arr) == 0 {
		return nil, errors.New("taskIds cannot be empty")
	}
	ids := make([]string, 0, len(arr))
	for _, raw := range arr {
		s, _ := raw.(string)
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		// backend supports short identifier in GET /tasks/:id, but bulk endpoints require full UUID
		t, err := client.GetTask(s)
		if err != nil {
			return nil, err
		}
		ids = append(ids, t.ID.String())
	}
	if len(ids) == 0 {
		return nil, errors.New("no valid task ids")
	}
	return ids, nil
}

func resolveProjectID(client *api.Client, projectIdentifier string) (string, error) {
	projectIdentifier = strings.TrimSpace(projectIdentifier)
	if projectIdentifier == "" {
		cfg, err := config.LoadConfig()
		if err == nil && cfg.ActiveProjectID != "" {
			return cfg.ActiveProjectID, nil
		}
		projects, err := client.ListProjects()
		if err != nil {
			return "", err
		}
		for _, p := range projects {
			if p.IsActive {
				return p.ID.String(), nil
			}
		}
		return "", errors.New("no active project")
	}

	if _, err := uuid.Parse(projectIdentifier); err == nil {
		return projectIdentifier, nil
	}

	projects, err := client.ListProjects()
	if err != nil {
		return "", err
	}
	for _, p := range projects {
		if p.Name == projectIdentifier || strings.HasPrefix(p.ID.String(), projectIdentifier) {
			return p.ID.String(), nil
		}
	}

	return "", errors.New("project not found")
}

func normalizePriority(s string) string {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return "M"
	}
	switch s {
	case "H", "HIGH":
		return "H"
	case "M", "MEDIUM":
		return "M"
	case "L", "LOW":
		return "L"
	default:
		return "M"
	}
}

func toInt(v interface{}) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	case string:
		var x int
		_, _ = fmt.Sscanf(t, "%d", &x)
		return x
	default:
		return 0
	}
}
