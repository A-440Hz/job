package handler

import (
	"net/http"
	"os"
	"path/filepath"
)

// ServeResume serves the resume PDF file from the backend
func (h *Handler) ServeResume(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow GET requests for actual file download
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resumePath := filepath.Join("internal", "static", "Haotian Zeng_Resume.pdf")
	if _, err := os.Stat(resumePath); os.IsNotExist(err) {
		http.Error(w, "Resume not found", http.StatusNotFound)
		return
	}

	fileBytes, err := os.ReadFile(resumePath)
	if err != nil {
		http.Error(w, "Error reading resume file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"Haotian Zeng_Resume.pdf\"")
	w.WriteHeader(http.StatusOK)
	w.Write(fileBytes)
}
