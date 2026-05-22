package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

//go:embed index.html
var indexHTML string

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/convert", handleConvert)

	log.Printf("md2pdf starting on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write([]byte(indexHTML))
}

func handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	markdown, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "body too large (max 10MB)", http.StatusRequestEntityTooLarge)
		return
	}

	if len(bytes.TrimSpace(markdown)) == 0 {
		http.Error(w, "empty markdown", http.StatusBadRequest)
		return
	}

	tmpDir, err := os.MkdirTemp("", "md2pdf-")
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)

	mdPath := filepath.Join(tmpDir, "input.md")
	pdfPath := filepath.Join(tmpDir, "output.pdf")

	if err := os.WriteFile(mdPath, markdown, 0644); err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	cmd := exec.CommandContext(ctx, "pandoc",
		mdPath,
		"-o", pdfPath,
		"--pdf-engine=wkhtmltopdf",
		"--metadata", "title=Document",
		"-V", "margin-top=15",
		"-V", "margin-bottom=15",
		"-V", "margin-left=15",
		"-V", "margin-right=15",
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	start := time.Now()
	if err := cmd.Run(); err != nil {
		log.Printf("pandoc failed: %v — stderr: %s", err, stderr.String())
		http.Error(w, "PDF conversion failed", http.StatusInternalServerError)
		return
	}
	log.Printf("converted %d bytes md → pdf in %v", len(markdown), time.Since(start))

	pdf, err := os.ReadFile(pdfPath)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"document.pdf\"")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdf)))
	w.Write(pdf)
}
