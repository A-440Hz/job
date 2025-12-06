package handler

import (
	_ "embed"
	"log"
	"net/http"
)

//go:embed Resume.pdf
var resume []byte

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

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"Haotian Zeng_Resume.pdf\"")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(resume); err != nil {
		log.Printf("Error writing resume to response: %v", err)
	}
}
