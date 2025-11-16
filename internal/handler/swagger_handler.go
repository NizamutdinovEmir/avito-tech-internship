package handler

import (
	"embed"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
)

//go:embed swagger-ui/*
var swaggerUI embed.FS

//go:embed openapi.yml
var openAPISpec embed.FS

const indexHTML = "index.html"

// ServeSwaggerUI serves Swagger UI
func ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Serve OpenAPI spec
	if path == "/swagger/openapi.yml" || path == "/swagger/openapi.yaml" {
		spec, err := openAPISpec.ReadFile("openapi.yml")
		if err != nil {
			http.Error(w, "OpenAPI spec not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/x-yaml")
		if _, err := w.Write(spec); err != nil {
			slog.Error("Failed to write OpenAPI spec", "error", err)
		}
		return
	}

	subFS, err := fs.Sub(swaggerUI, "swagger-ui")
	if err != nil {
		http.Error(w, "Swagger UI not found", http.StatusNotFound)
		return
	}

	filePath := path[len("/swagger"):]
	if filePath == "" || filePath == "/" {
		filePath = indexHTML
	} else {
		if len(filePath) > 0 && filePath[0] == '/' {
			filePath = filePath[1:]
		}
	}

	file, err := subFS.Open(filePath)
	if err != nil {
		if filePath != indexHTML {
			filePath = indexHTML
			file, err = subFS.Open(filePath)
		}
		if err != nil {
			http.Error(w, "File not found: "+filePath, http.StatusNotFound)
			return
		}
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	if filePath == indexHTML {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	} else if len(filePath) > 4 && filePath[len(filePath)-4:] == ".css" {
		w.Header().Set("Content-Type", "text/css")
	} else if len(filePath) > 3 && filePath[len(filePath)-3:] == ".js" {
		w.Header().Set("Content-Type", "application/javascript")
	}

	if _, err := w.Write(content); err != nil {
		slog.Error("Failed to write file content", "error", err)
	}
}
