// Package main implements an MCP (Model Context Protocol) server that provides
// AI agents with access to the Dailies task management API.
//
// The server exposes 15 tools across three categories:
//   - Task management: list, get, create, update, delete tasks
//   - Tag management: list, get, create, update, delete tags
//   - Frequency management: list, get, create, update, delete frequencies
//
// Usage:
//
//	./mcp [--host hostname:port]
//
// The server runs on stdio transport and accepts a --host flag to specify
// the Dailies API endpoint (defaults to localhost:9001).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jhoffmann/dailies/internal/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DailiesClient provides an HTTP client interface to the Dailies API.
type DailiesClient struct {
	baseURL string
	client  *http.Client
}

// NewDailiesClient creates a new Dailies API client with the specified base URL.
func NewDailiesClient(baseURL string) *DailiesClient {
	return &DailiesClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// makeRequest performs an HTTP request to the Dailies API and returns the response body.
// It handles JSON marshaling of the request body and error handling for HTTP status codes.
func (d *DailiesClient) makeRequest(method, endpoint string, body any) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, d.baseURL+endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Request types for MCP tool arguments. These are lightweight DTOs for input validation.
// Response types are also DTOs to avoid circular dependencies in JSON schema generation.

// ListTasksRequest represents the parameters for filtering and sorting tasks.
type ListTasksRequest struct {
	Completed *bool  `json:"completed,omitempty" jsonschema:"Filter by completion status"`
	Name      string `json:"name,omitempty" jsonschema:"Filter by task name (partial matching)"`
	TagIDs    string `json:"tag_ids,omitempty" jsonschema:"Comma-separated tag IDs"`
	Sort      string `json:"sort,omitempty" jsonschema:"Sort field: completed, priority, name"`
}

// GetByIDRequest represents a request to retrieve a resource by its UUID.
type GetByIDRequest struct {
	ID string `json:"id" jsonschema:"UUID of the resource"`
}

// CreateTaskRequest represents the parameters for creating a new task.
type CreateTaskRequest struct {
	Name        string   `json:"name" jsonschema:"Task name"`
	FrequencyID *string  `json:"frequency_id,omitempty" jsonschema:"Optional frequency UUID"`
	TagIDs      []string `json:"tag_ids,omitempty" jsonschema:"Array of tag UUIDs"`
}

// UpdateTaskRequest represents the parameters for updating an existing task.
type UpdateTaskRequest struct {
	ID          string  `json:"id" jsonschema:"Task UUID"`
	Name        string  `json:"name,omitempty" jsonschema:"Task name"`
	Completed   *bool   `json:"completed,omitempty" jsonschema:"Completion status"`
	Priority    *int    `json:"priority,omitempty" jsonschema:"Priority (1-5)"`
	FrequencyID *string `json:"frequency_id,omitempty" jsonschema:"Frequency UUID"`
}

// ListTagsRequest represents the parameters for filtering tags.
type ListTagsRequest struct {
	Name string `json:"name,omitempty" jsonschema:"Filter by tag name (partial matching)"`
}

// CreateTagRequest represents the parameters for creating a new tag.
type CreateTagRequest struct {
	Name  string `json:"name" jsonschema:"Tag name"`
	Color string `json:"color,omitempty" jsonschema:"Hex color code (auto-generated if not provided)"`
}

// UpdateTagRequest represents the parameters for updating an existing tag.
type UpdateTagRequest struct {
	ID    string `json:"id" jsonschema:"Tag UUID"`
	Name  string `json:"name,omitempty" jsonschema:"Tag name"`
	Color string `json:"color,omitempty" jsonschema:"Hex color code"`
}

// ListFrequenciesRequest represents the parameters for filtering frequencies.
type ListFrequenciesRequest struct {
	Name string `json:"name,omitempty" jsonschema:"Filter by frequency name (partial matching)"`
}

// CreateFrequencyRequest represents the parameters for creating a new frequency.
type CreateFrequencyRequest struct {
	Name  string `json:"name" jsonschema:"Frequency name"`
	Reset string `json:"reset" jsonschema:"Cron expression for reset timing"`
}

// UpdateFrequencyRequest represents the parameters for updating an existing frequency.
type UpdateFrequencyRequest struct {
	ID    string `json:"id" jsonschema:"Frequency UUID"`
	Name  string `json:"name,omitempty" jsonschema:"Frequency name"`
	Reset string `json:"reset,omitempty" jsonschema:"Cron expression for reset timing"`
}

// SuccessMessage represents a successful operation response.
type SuccessMessage struct {
	Message string `json:"message"`
}

// Response DTOs to avoid circular dependencies in JSON schema generation.
// These mirror the models but without back-references.

// TaskResponse represents a task without circular references.
type TaskResponse struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	DateCreated  string             `json:"date_created"`
	DateModified string             `json:"date_modified"`
	Completed    bool               `json:"completed"`
	Priority     int                `json:"priority"`
	FrequencyID  *string            `json:"frequency_id,omitempty"`
	Frequency    *FrequencyResponse `json:"frequency,omitempty"`
	Tags         []TagResponse      `json:"tags,omitempty"`
}

// TagResponse represents a tag without circular references.
type TagResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// FrequencyResponse represents a frequency without circular references.
type FrequencyResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Reset string `json:"reset"`
}

// List response wrappers to satisfy MCP requirement for object-type responses

// TaskListResponse wraps an array of tasks in an object.
type TaskListResponse struct {
	Tasks []TaskResponse `json:"tasks"`
}

// TagListResponse wraps an array of tags in an object.
type TagListResponse struct {
	Tags []TagResponse `json:"tags"`
}

// FrequencyListResponse wraps an array of frequencies in an object.
type FrequencyListResponse struct {
	Frequencies []FrequencyResponse `json:"frequencies"`
}

// Conversion functions from models to response DTOs

// taskToResponse converts a models.Task to TaskResponse.
func taskToResponse(task models.Task) TaskResponse {
	response := TaskResponse{
		ID:           task.ID.String(),
		Name:         task.Name,
		DateCreated:  task.DateCreated.Format(time.RFC3339),
		DateModified: task.DateModified.Format(time.RFC3339),
		Completed:    task.Completed,
		Priority:     task.Priority,
	}

	if task.FrequencyID != nil {
		freqID := task.FrequencyID.String()
		response.FrequencyID = &freqID
	}

	if task.Frequency != nil {
		response.Frequency = &FrequencyResponse{
			ID:    task.Frequency.ID.String(),
			Name:  task.Frequency.Name,
			Reset: task.Frequency.Reset,
		}
	}

	for _, tag := range task.Tags {
		response.Tags = append(response.Tags, TagResponse{
			ID:    tag.ID.String(),
			Name:  tag.Name,
			Color: tag.Color,
		})
	}

	return response
}

// tagToResponse converts a models.Tag to TagResponse.
func tagToResponse(tag models.Tag) TagResponse {
	return TagResponse{
		ID:    tag.ID.String(),
		Name:  tag.Name,
		Color: tag.Color,
	}
}

// frequencyToResponse converts a models.Frequency to FrequencyResponse.
func frequencyToResponse(freq models.Frequency) FrequencyResponse {
	return FrequencyResponse{
		ID:    freq.ID.String(),
		Name:  freq.Name,
		Reset: freq.Reset,
	}
}

// main initializes and runs the MCP server that provides AI agents access to the Dailies API.
// It accepts a --host flag to specify the API endpoint (defaults to localhost:9001).
// The server runs on stdio transport and registers tools for task, tag, and frequency management.
func main() {
	host := flag.String("host", "localhost:9001", "API host in the form hostname:port")
	flag.Parse()

	// Parse and validate host
	apiURL, err := parseHost(*host)
	if err != nil {
		log.Fatalf("Invalid host format: %v", err)
	}

	log.Printf("Starting MCP server, connecting to Dailies API at %s", apiURL)

	// Initialize dailies client
	dailiesClient := NewDailiesClient(apiURL)

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "dailies-mcp",
		Version: "1.0.0",
	}, nil)

	// Register tools
	registerTaskTools(server, dailiesClient)
	registerTagTools(server, dailiesClient)
	registerFrequencyTools(server, dailiesClient)

	// Start MCP server on stdio transport
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

// registerTaskTools adds task management tools to the MCP server.
// Provides tools for listing, getting, creating, updating, and deleting tasks.
// All response types use TaskResponse DTOs to avoid circular dependencies.
func registerTaskTools(server *mcp.Server, client *DailiesClient) {
	// List tasks
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tasks",
		Description: "List all tasks with optional filtering by completion status, name, or tags",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListTasksRequest) (*mcp.CallToolResult, TaskListResponse, error) {
		params := url.Values{}
		if args.Completed != nil {
			params.Set("completed", fmt.Sprintf("%v", *args.Completed))
		}
		if args.Name != "" {
			params.Set("name", args.Name)
		}
		if args.TagIDs != "" {
			params.Set("tag_ids", args.TagIDs)
		}
		if args.Sort != "" {
			params.Set("sort", args.Sort)
		}

		endpoint := "/tasks"
		if len(params) > 0 {
			endpoint += "?" + params.Encode()
		}

		respBody, err := client.makeRequest("GET", endpoint, nil)
		if err != nil {
			return nil, TaskListResponse{}, err
		}

		var tasks []models.Task
		if err := json.Unmarshal(respBody, &tasks); err != nil {
			return nil, TaskListResponse{}, fmt.Errorf("failed to parse tasks response: %w", err)
		}

		// Convert to response DTOs
		var responseTasks []TaskResponse
		for _, task := range tasks {
			responseTasks = append(responseTasks, taskToResponse(task))
		}

		return nil, TaskListResponse{Tasks: responseTasks}, nil
	})

	// Get task by ID
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_task",
		Description: "Get a specific task by ID with full details including tags and frequency",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetByIDRequest) (*mcp.CallToolResult, TaskResponse, error) {
		respBody, err := client.makeRequest("GET", "/tasks/"+args.ID, nil)
		if err != nil {
			return nil, TaskResponse{}, err
		}

		var task models.Task
		if err := json.Unmarshal(respBody, &task); err != nil {
			return nil, TaskResponse{}, fmt.Errorf("failed to parse task response: %w", err)
		}

		return nil, taskToResponse(task), nil
	})

	// Create task
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_task",
		Description: "Create a new task with optional frequency assignment and tag associations",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateTaskRequest) (*mcp.CallToolResult, TaskResponse, error) {
		respBody, err := client.makeRequest("POST", "/tasks", args)
		if err != nil {
			return nil, TaskResponse{}, err
		}

		var task models.Task
		if err := json.Unmarshal(respBody, &task); err != nil {
			return nil, TaskResponse{}, fmt.Errorf("failed to parse task response: %w", err)
		}

		return nil, taskToResponse(task), nil
	})

	// Update task
	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_task",
		Description: "Update an existing task's name, completion status, priority, or frequency",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateTaskRequest) (*mcp.CallToolResult, TaskResponse, error) {
		id := args.ID
		// Remove ID from request body - it goes in the URL path
		updateData := map[string]interface{}{}
		if args.Name != "" {
			updateData["name"] = args.Name
		}
		if args.Completed != nil {
			updateData["completed"] = *args.Completed
		}
		if args.Priority != nil {
			updateData["priority"] = *args.Priority
		}
		if args.FrequencyID != nil {
			updateData["frequency_id"] = *args.FrequencyID
		}

		respBody, err := client.makeRequest("PUT", "/tasks/"+id, updateData)
		if err != nil {
			return nil, TaskResponse{}, err
		}

		var task models.Task
		if err := json.Unmarshal(respBody, &task); err != nil {
			return nil, TaskResponse{}, fmt.Errorf("failed to parse task response: %w", err)
		}

		return nil, taskToResponse(task), nil
	})

	// Delete task
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_task",
		Description: "Delete a task by ID - this action cannot be undone",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetByIDRequest) (*mcp.CallToolResult, SuccessMessage, error) {
		_, err := client.makeRequest("DELETE", "/tasks/"+args.ID, nil)
		if err != nil {
			return nil, SuccessMessage{}, err
		}

		return nil, SuccessMessage{Message: "Task deleted successfully"}, nil
	})
}

// registerTagTools adds tag management tools to the MCP server.
// Provides tools for listing, getting, creating, updating, and deleting tags.
// All response types use TagResponse DTOs to avoid circular dependencies.
func registerTagTools(server *mcp.Server, client *DailiesClient) {
	// List tags
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tags",
		Description: "List all tags with optional name filtering",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListTagsRequest) (*mcp.CallToolResult, TagListResponse, error) {
		params := url.Values{}
		if args.Name != "" {
			params.Set("name", args.Name)
		}

		endpoint := "/tags"
		if len(params) > 0 {
			endpoint += "?" + params.Encode()
		}

		respBody, err := client.makeRequest("GET", endpoint, nil)
		if err != nil {
			return nil, TagListResponse{}, err
		}

		var tags []models.Tag
		if err := json.Unmarshal(respBody, &tags); err != nil {
			return nil, TagListResponse{}, fmt.Errorf("failed to parse tags response: %w", err)
		}

		// Convert to response DTOs
		var responseTags []TagResponse
		for _, tag := range tags {
			responseTags = append(responseTags, tagToResponse(tag))
		}

		return nil, TagListResponse{Tags: responseTags}, nil
	})

	// Get tag by ID
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_tag",
		Description: "Get a specific tag by ID",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetByIDRequest) (*mcp.CallToolResult, TagResponse, error) {
		respBody, err := client.makeRequest("GET", "/tags/"+args.ID, nil)
		if err != nil {
			return nil, TagResponse{}, err
		}

		var tag models.Tag
		if err := json.Unmarshal(respBody, &tag); err != nil {
			return nil, TagResponse{}, fmt.Errorf("failed to parse tag response: %w", err)
		}

		return nil, tagToResponse(tag), nil
	})

	// Create tag
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_tag",
		Description: "Create a new tag with optional color (auto-generated if not provided)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateTagRequest) (*mcp.CallToolResult, TagResponse, error) {
		respBody, err := client.makeRequest("POST", "/tags", args)
		if err != nil {
			return nil, TagResponse{}, err
		}

		var tag models.Tag
		if err := json.Unmarshal(respBody, &tag); err != nil {
			return nil, TagResponse{}, fmt.Errorf("failed to parse tag response: %w", err)
		}

		return nil, tagToResponse(tag), nil
	})

	// Update tag
	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_tag",
		Description: "Update an existing tag's name and/or color",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateTagRequest) (*mcp.CallToolResult, TagResponse, error) {
		id := args.ID
		updateData := map[string]interface{}{}
		if args.Name != "" {
			updateData["name"] = args.Name
		}
		if args.Color != "" {
			updateData["color"] = args.Color
		}

		respBody, err := client.makeRequest("PUT", "/tags/"+id, updateData)
		if err != nil {
			return nil, TagResponse{}, err
		}

		var tag models.Tag
		if err := json.Unmarshal(respBody, &tag); err != nil {
			return nil, TagResponse{}, fmt.Errorf("failed to parse tag response: %w", err)
		}

		return nil, tagToResponse(tag), nil
	})

	// Delete tag
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_tag",
		Description: "Delete a tag by ID - this will remove the tag from all associated tasks",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetByIDRequest) (*mcp.CallToolResult, SuccessMessage, error) {
		_, err := client.makeRequest("DELETE", "/tags/"+args.ID, nil)
		if err != nil {
			return nil, SuccessMessage{}, err
		}

		return nil, SuccessMessage{Message: "Tag deleted successfully"}, nil
	})
}

// registerFrequencyTools adds frequency management tools to the MCP server.
// Provides tools for listing, getting, creating, updating, and deleting frequencies.
// All response types use FrequencyResponse DTOs to avoid circular dependencies.
func registerFrequencyTools(server *mcp.Server, client *DailiesClient) {
	// List frequencies
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_frequencies",
		Description: "List all frequencies with optional name filtering",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ListFrequenciesRequest) (*mcp.CallToolResult, FrequencyListResponse, error) {
		params := url.Values{}
		if args.Name != "" {
			params.Set("name", args.Name)
		}

		endpoint := "/frequencies"
		if len(params) > 0 {
			endpoint += "?" + params.Encode()
		}

		respBody, err := client.makeRequest("GET", endpoint, nil)
		if err != nil {
			return nil, FrequencyListResponse{}, err
		}

		var frequencies []models.Frequency
		if err := json.Unmarshal(respBody, &frequencies); err != nil {
			return nil, FrequencyListResponse{}, fmt.Errorf("failed to parse frequencies response: %w", err)
		}

		// Convert to response DTOs
		var responseFrequencies []FrequencyResponse
		for _, frequency := range frequencies {
			responseFrequencies = append(responseFrequencies, frequencyToResponse(frequency))
		}

		return nil, FrequencyListResponse{Frequencies: responseFrequencies}, nil
	})

	// Get frequency by ID
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_frequency",
		Description: "Get a specific frequency by ID with its cron schedule",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetByIDRequest) (*mcp.CallToolResult, FrequencyResponse, error) {
		respBody, err := client.makeRequest("GET", "/frequencies/"+args.ID, nil)
		if err != nil {
			return nil, FrequencyResponse{}, err
		}

		var frequency models.Frequency
		if err := json.Unmarshal(respBody, &frequency); err != nil {
			return nil, FrequencyResponse{}, fmt.Errorf("failed to parse frequency response: %w", err)
		}

		return nil, frequencyToResponse(frequency), nil
	})

	// Create frequency
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_frequency",
		Description: "Create a new frequency with cron expression for automatic task resets",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args CreateFrequencyRequest) (*mcp.CallToolResult, FrequencyResponse, error) {
		respBody, err := client.makeRequest("POST", "/frequencies", args)
		if err != nil {
			return nil, FrequencyResponse{}, err
		}

		var frequency models.Frequency
		if err := json.Unmarshal(respBody, &frequency); err != nil {
			return nil, FrequencyResponse{}, fmt.Errorf("failed to parse frequency response: %w", err)
		}

		return nil, frequencyToResponse(frequency), nil
	})

	// Update frequency
	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_frequency",
		Description: "Update an existing frequency's name and/or cron schedule",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args UpdateFrequencyRequest) (*mcp.CallToolResult, FrequencyResponse, error) {
		id := args.ID
		updateData := map[string]interface{}{}
		if args.Name != "" {
			updateData["name"] = args.Name
		}
		if args.Reset != "" {
			updateData["reset"] = args.Reset
		}

		respBody, err := client.makeRequest("PUT", "/frequencies/"+id, updateData)
		if err != nil {
			return nil, FrequencyResponse{}, err
		}

		var frequency models.Frequency
		if err := json.Unmarshal(respBody, &frequency); err != nil {
			return nil, FrequencyResponse{}, fmt.Errorf("failed to parse frequency response: %w", err)
		}

		return nil, frequencyToResponse(frequency), nil
	})

	// Delete frequency
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_frequency",
		Description: "Delete a frequency by ID - tasks using this frequency will have their frequency cleared",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args GetByIDRequest) (*mcp.CallToolResult, SuccessMessage, error) {
		_, err := client.makeRequest("DELETE", "/frequencies/"+args.ID, nil)
		if err != nil {
			return nil, SuccessMessage{}, err
		}

		return nil, SuccessMessage{Message: "Frequency deleted successfully"}, nil
	})
}

// parseHost parses the host flag and returns a complete API URL.
// It handles default port (9001) and localhost hostname when not specified.
// Examples:
//   - "api.example.com" becomes "http://api.example.com:9001"
//   - ":8080" becomes "http://localhost:8080"
//   - "localhost:9001" becomes "http://localhost:9001"
func parseHost(host string) (string, error) {
	// If no port is specified, add default port
	if !strings.Contains(host, ":") {
		host = host + ":9001"
	}

	// If it starts with :, prepend localhost
	if strings.HasPrefix(host, ":") {
		host = "localhost" + host
	}

	// Validate by parsing as URL
	u, err := url.Parse("http://" + host)
	if err != nil {
		return "", fmt.Errorf("failed to parse host: %w", err)
	}

	return u.String(), nil
}
