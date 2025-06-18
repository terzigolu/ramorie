package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/terzigolu/josepshbrain-go/internal/models"
)

// --- Notes Endpoints ---

// ListNotes handles GET /tasks/{taskId}/notes
func ListNotes(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement fetching notes from DB
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AnnotationListResponse{
		Success: true,
		Data:    []models.Annotation{},
	})
}

// AddNote handles POST /tasks/{taskId}/notes
func AddNote(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement adding a note to DB
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true})
}

// UpdateNote handles PUT /tasks/{taskId}/notes/{noteId}
func UpdateNote(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement updating a note in DB
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true})
}

// DeleteNote handles DELETE /tasks/{taskId}/notes/{noteId}
func DeleteNote(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement deleting a note from DB
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true})
}

// --- AI Feature Endpoints ---

// AIAnalysis handles POST /tasks/{taskId}/ai/analysis
func AIAnalysis(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement AI analysis logic
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Data: "AI analysis result"})
}

// AIDecompose handles POST /tasks/{taskId}/ai/decompose
func AIDecompose(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement AI decompose logic
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Data: []string{}})
}

// AIPriority handles POST /tasks/{taskId}/ai/priority
func AIPriority(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement AI priority suggestion logic
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Data: map[string]string{"suggestion": "M", "explanation": "AI explanation"}})
}

// AIElaborate handles POST /tasks/{taskId}/ai/elaborate
func AIElaborate(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement AI elaborate logic
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Data: "Elaborated task description"})
}

// AISuggestions handles POST /tasks/{taskId}/ai/suggestions
func AISuggestions(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement AI suggestions logic
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Data: []string{}})
}