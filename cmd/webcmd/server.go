// Package webcmd provides a web-based configuration UI for protobuild.
package webcmd

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/pubgo/protobuild/internal/config"
)

//go:embed templates/*
var templateFS embed.FS

// CommandResult represents the result of a command execution.
type CommandResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

// Server represents the web server.
type Server struct {
	configPath string
	config     *config.Config
	mu         sync.RWMutex
	server     *http.Server
	templates  *template.Template
}

// NewServer creates a new web server.
func NewServer(configPath string) (*Server, error) {
	s := &Server{
		configPath: configPath,
	}

	// Load templates
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}
	s.templates = tmpl

	// Load initial config
	if err := s.loadConfig(); err != nil {
		// Create default config if not exists
		s.config = config.Default()
	}

	return s, nil
}

// loadConfig loads the configuration from file.
func (s *Server) loadConfig() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := config.Load(s.configPath)
	if err != nil {
		return err
	}

	s.config = cfg
	return nil
}

// saveConfig saves the configuration to file.
func (s *Server) saveConfig() error {
	s.mu.RLock()
	cfg := s.config
	s.mu.RUnlock()

	return config.Save(s.configPath, cfg)
}

// Start starts the web server.
func (s *Server) Start(ctx context.Context, port int) error {
	mux := http.NewServeMux()

	// Static files
	mux.HandleFunc("/", s.handleIndex)

	// API endpoints
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/config/save", s.handleSaveConfig)
	mux.HandleFunc("/api/command/", s.handleCommand)
	mux.HandleFunc("/api/command-stream/", s.handleCommandStream)
	mux.HandleFunc("/api/proto-files", s.handleProtoFiles)
	mux.HandleFunc("/api/proto-content", s.handleProtoContent)
	mux.HandleFunc("/api/deps/status", s.handleDepsStatus)
	mux.HandleFunc("/api/project/stats", s.handleProjectStats)

	addr := fmt.Sprintf(":%d", port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Find available port if default is in use
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		// Try to find an available port
		listener, err = net.Listen("tcp", ":0")
		if err != nil {
			return fmt.Errorf("failed to find available port: %w", err)
		}
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port
	url := fmt.Sprintf("http://localhost:%d", actualPort)

	slog.Info("Starting web server", "url", url)
	fmt.Printf("\n🌐 Web UI available at: %s\n\n", url)

	// Open browser
	go func() {
		time.Sleep(500 * time.Millisecond)
		openBrowser(url)
	}()

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.server.Shutdown(shutdownCtx)
	}()

	return s.server.Serve(listener)
}

// handleIndex serves the main page.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	s.mu.RLock()
	cfg := s.config
	s.mu.RUnlock()

	data := map[string]interface{}{
		"Config":     cfg,
		"ConfigPath": s.configPath,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleConfig returns the current configuration.
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Reload config from file
		s.loadConfig()

		s.mu.RLock()
		cfg := s.config
		s.mu.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleSaveConfig saves the configuration.
func (s *Server) handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.config = &cfg
	s.mu.Unlock()

	if err := s.saveConfig(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CommandResult{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CommandResult{
		Success: true,
		Output:  "Configuration saved successfully",
	})
}

// handleCommand executes protobuild commands.
func (s *Server) handleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract command from URL path
	cmdName := strings.TrimPrefix(r.URL.Path, "/api/command/")
	if cmdName == "" {
		http.Error(w, "Command name required", http.StatusBadRequest)
		return
	}

	// Build command arguments
	args := []string{"-c", s.configPath, cmdName}

	// Parse additional flags from request body
	var flags map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&flags); err == nil {
		for key, val := range flags {
			switch v := val.(type) {
			case bool:
				if v {
					args = append(args, "--"+key)
				}
			case string:
				if v != "" {
					args = append(args, "--"+key, v)
				}
			}
		}
	}

	// Get executable path
	executable, err := os.Executable()
	if err != nil {
		executable = "protobuild"
	}

	// Execute command
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = filepath.Dir(s.configPath)

	output, err := cmd.CombinedOutput()

	result := CommandResult{
		Success: err == nil,
		Output:  string(output),
	}
	if err != nil {
		result.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleProtoFiles returns a list of proto files in the project.
func (s *Server) handleProtoFiles(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	cfg := s.config
	s.mu.RUnlock()

	var files []string
	baseDir := filepath.Dir(s.configPath)

	for _, root := range cfg.Root {
		rootPath := filepath.Join(baseDir, root)
		filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(path, ".proto") {
				relPath, _ := filepath.Rel(baseDir, path)
				files = append(files, relPath)
			}
			return nil
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// handleDepsStatus returns the status of dependencies.
func (s *Server) handleDepsStatus(w http.ResponseWriter, r *http.Request) {
	// Get executable path
	executable, err := os.Executable()
	if err != nil {
		executable = "protobuild"
	}

	cmd := exec.Command(executable, "-c", s.configPath, "deps")
	cmd.Dir = filepath.Dir(s.configPath)
	output, _ := cmd.CombinedOutput()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"output": string(output),
	})
}

// openBrowser opens the URL in the default browser.
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		return fmt.Errorf("unsupported platform")
	}

	return exec.Command(cmd, args...).Start()
}

// handleCommandStream executes a command and streams output via SSE.
func (s *Server) handleCommandStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract command from URL path
	cmdName := strings.TrimPrefix(r.URL.Path, "/api/command-stream/")
	if cmdName == "" {
		http.Error(w, "Command name required", http.StatusBadRequest)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Build command arguments
	args := []string{"-c", s.configPath, cmdName}

	// Get executable path
	executable, err := os.Executable()
	if err != nil {
		executable = "protobuild"
	}

	// Execute command with streaming output
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = filepath.Dir(s.configPath)

	// Create pipes for stdout and stderr
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(w, "data: {\"type\":\"error\",\"data\":\"%s\"}\n\n", err.Error())
		flusher.Flush()
		return
	}

	// Stream output
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Fprintf(w, "data: {\"type\":\"stdout\",\"data\":\"%s\"}\n\n", escapeJSON(line))
			flusher.Flush()
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Fprintf(w, "data: {\"type\":\"stderr\",\"data\":\"%s\"}\n\n", escapeJSON(line))
			flusher.Flush()
		}
	}()

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintf(w, "data: {\"type\":\"error\",\"data\":\"%s\"}\n\n", err.Error())
	} else {
		fmt.Fprintf(w, "data: {\"type\":\"done\",\"data\":\"Command completed successfully\"}\n\n")
	}
	flusher.Flush()
}

// escapeJSON escapes special characters for JSON string.
func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// handleProtoContent returns the content of a specific proto file.
func (s *Server) handleProtoContent(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("file")
	if filePath == "" {
		http.Error(w, "File path required", http.StatusBadRequest)
		return
	}

	baseDir := filepath.Dir(s.configPath)
	fullPath := filepath.Join(baseDir, filePath)

	// Security check: ensure the path is within the project
	absBase, _ := filepath.Abs(baseDir)
	absPath, _ := filepath.Abs(fullPath)
	if !strings.HasPrefix(absPath, absBase) {
		http.Error(w, "Invalid file path", http.StatusForbidden)
		return
	}

	// Check file extension
	if !strings.HasSuffix(fullPath, ".proto") {
		http.Error(w, "Only .proto files allowed", http.StatusForbidden)
		return
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Get file info
	info, _ := os.Stat(fullPath)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"path":     filePath,
		"content":  string(content),
		"size":     info.Size(),
		"modified": info.ModTime().Format(time.RFC3339),
	})
}

// ProjectStats represents project statistics.
type ProjectStats struct {
	ProtoFiles     int      `json:"proto_files"`
	TotalLines     int      `json:"total_lines"`
	MessageCount   int      `json:"message_count"`
	ServiceCount   int      `json:"service_count"`
	DependencyCount int     `json:"dependency_count"`
	PluginCount    int      `json:"plugin_count"`
	ProtoRoots     []string `json:"proto_roots"`
	VendorDir      string   `json:"vendor_dir"`
	VendorFiles    int      `json:"vendor_files"`
}

// handleProjectStats returns project statistics.
func (s *Server) handleProjectStats(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	cfg := s.config
	s.mu.RUnlock()

	stats := ProjectStats{
		ProtoRoots:      cfg.Root,
		VendorDir:       cfg.Vendor,
		DependencyCount: len(cfg.Depends),
		PluginCount:     len(cfg.Plugins),
	}

	baseDir := filepath.Dir(s.configPath)

	// Count proto files in root directories
	for _, root := range cfg.Root {
		rootPath := filepath.Join(baseDir, root)
		filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(path, ".proto") {
				stats.ProtoFiles++

				// Count lines, messages, and services
				content, err := os.ReadFile(path)
				if err == nil {
					lines := strings.Split(string(content), "\n")
					stats.TotalLines += len(lines)

					for _, line := range lines {
						trimmed := strings.TrimSpace(line)
						if strings.HasPrefix(trimmed, "message ") {
							stats.MessageCount++
						} else if strings.HasPrefix(trimmed, "service ") {
							stats.ServiceCount++
						}
					}
				}
			}
			return nil
		})
	}

	// Count vendor files
	if cfg.Vendor != "" {
		vendorPath := filepath.Join(baseDir, cfg.Vendor)
		filepath.Walk(vendorPath, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(path, ".proto") {
				stats.VendorFiles++
			}
			return nil
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
