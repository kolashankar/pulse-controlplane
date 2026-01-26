package services

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// DeveloperToolsService handles SDK generation and documentation
type DeveloperToolsService struct{}

// NewDeveloperToolsService creates a new developer tools service
func NewDeveloperToolsService() *DeveloperToolsService {
	return &DeveloperToolsService{}
}

// PostmanCollection represents a Postman collection structure
type PostmanCollection struct {
	Info struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Schema      string `json:"schema"`
		Version     string `json:"version"`
	} `json:"info"`
	Item []PostmanItem `json:"item"`
	Variable []PostmanVariable `json:"variable"`
}

type PostmanItem struct {
	Name    string         `json:"name"`
	Request PostmanRequest `json:"request,omitempty"`
	Item    []PostmanItem  `json:"item,omitempty"`
}

type PostmanRequest struct {
	Method string          `json:"method"`
	Header []PostmanHeader `json:"header"`
	URL    PostmanURL      `json:"url"`
	Body   *PostmanBody    `json:"body,omitempty"`
}

type PostmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PostmanURL struct {
	Raw  string   `json:"raw"`
	Host []string `json:"host"`
	Path []string `json:"path"`
}

type PostmanBody struct {
	Mode string `json:"mode"`
	Raw  string `json:"raw,omitempty"`
}

type PostmanVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GeneratePostmanCollection generates a Postman collection
func (s *DeveloperToolsService) GeneratePostmanCollection() (*PostmanCollection, error) {
	collection := &PostmanCollection{}
	collection.Info.Name = "Pulse Control Plane API"
	collection.Info.Description = "Complete API collection for Pulse Control Plane"
	collection.Info.Schema = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
	collection.Info.Version = "1.0.0"
	
	// Add environment variables
	collection.Variable = []PostmanVariable{
		{Key: "base_url", Value: "http://localhost:8081/api/v1"},
		{Key: "pulse_api_key", Value: "your_pulse_api_key"},
	}
	
	// Organizations
	orgItems := []PostmanItem{
		{
			Name: "Create Organization",
			Request: PostmanRequest{
				Method: "POST",
				URL: PostmanURL{
					Raw: "{{base_url}}/organizations",
					Host: []string{"{{base_url}}"},
					Path: []string{"organizations"},
				},
				Body: &PostmanBody{
					Mode: "raw",
					Raw: `{"name": "Example Org", "admin_email": "admin@example.com", "plan": "pro"}`,
				},
			},
		},
		{
			Name: "List Organizations",
			Request: PostmanRequest{
				Method: "GET",
				URL: PostmanURL{
					Raw: "{{base_url}}/organizations",
					Host: []string{"{{base_url}}"},
					Path: []string{"organizations"},
				},
			},
		},
	}
	
	// Projects
	projectItems := []PostmanItem{
		{
			Name: "Create Project",
			Request: PostmanRequest{
				Method: "POST",
				URL: PostmanURL{
					Raw: "{{base_url}}/projects",
					Host: []string{"{{base_url}}"},
					Path: []string{"projects"},
				},
				Body: &PostmanBody{
					Mode: "raw",
					Raw: `{"name": "My Project", "region": "us-east"}`,
				},
			},
		},
		{
			Name: "List Projects",
			Request: PostmanRequest{
				Method: "GET",
				URL: PostmanURL{
					Raw: "{{base_url}}/projects",
					Host: []string{"{{base_url}}"},
					Path: []string{"projects"},
				},
			},
		},
	}
	
	// Tokens
	tokenItems := []PostmanItem{
		{
			Name: "Create Token",
			Request: PostmanRequest{
				Method: "POST",
				Header: []PostmanHeader{
					{Key: "X-Pulse-Key", Value: "{{pulse_api_key}}"},
				},
				URL: PostmanURL{
					Raw: "{{base_url}}/tokens/create",
					Host: []string{"{{base_url}}"},
					Path: []string{"tokens", "create"},
				},
				Body: &PostmanBody{
					Mode: "raw",
					Raw: `{"room_name": "my-room", "identity": "user-123"}`,
				},
			},
		},
	}
	
	collection.Item = []PostmanItem{
		{Name: "Organizations", Item: orgItems},
		{Name: "Projects", Item: projectItems},
		{Name: "Tokens", Item: tokenItems},
	}
	
	return collection, nil
}

// GenerateSDK generates SDK for a specific language
func (s *DeveloperToolsService) GenerateSDK(language string) ([]byte, error) {
	// This is a placeholder implementation
	// In production, you would use OpenAPI Generator or similar tools
	
	switch language {
	case "go":
		return s.generateGoSDK()
	case "javascript":
		return s.generateJavaScriptSDK()
	case "python":
		return s.generatePythonSDK()
	default:
		return nil, fmt.Errorf("unsupported language: %s", language)
	}
}

func (s *DeveloperToolsService) generateGoSDK() ([]byte, error) {
	// Create a simple Go SDK package
	sdkCode := `package pulsesdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	BaseURL string
	APIKey  string
	HTTPClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) CreateToken(roomName, identity string) (map[string]interface{}, error) {
	body := map[string]string{
		"room_name": roomName,
		"identity": identity,
	}
	
	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", c.BaseURL+"/tokens/create", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Pulse-Key", c.APIKey)
	
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result, nil
}

// Add more methods as needed
`
	
	readmeCode := `# Pulse Go SDK

## Installation

` + "```bash" + `
go get github.com/pulse/go-sdk
` + "```" + `

## Usage

` + "```go" + `
package main

import (
	"fmt"
	"github.com/pulse/go-sdk"
)

func main() {
	client := pulsesdk.NewClient("http://localhost:8081/api/v1", "your_api_key")
	token, err := client.CreateToken("my-room", "user-123")
	if err != nil {
		panic(err)
	}
	fmt.Println("Token:", token)
}
` + "```" + `
`
	
	// Create a zip file with SDK contents
	return s.createZip(map[string]string{
		"pulse-sdk/client.go": sdkCode,
		"pulse-sdk/README.md": readmeCode,
	})
}

func (s *DeveloperToolsService) generateJavaScriptSDK() ([]byte, error) {
	sdkCode := `class PulseClient {
  constructor(baseURL, apiKey) {
    this.baseURL = baseURL;
    this.apiKey = apiKey;
  }

  async createToken(roomName, identity) {
    const response = await fetch(this.baseURL + '/tokens/create', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Pulse-Key': this.apiKey
      },
      body: JSON.stringify({ room_name: roomName, identity: identity })
    });
    
    if (!response.ok) {
      throw new Error('API error: ' + response.status);
    }
    
    return await response.json();
  }
  
  // Add more methods as needed
}

module.exports = PulseClient;
`
	
	readmeCode := `# Pulse JavaScript SDK

## Installation

` + "```bash" + `
npm install @pulse/sdk
` + "```" + `

## Usage

` + "```javascript" + `
const PulseClient = require('@pulse/sdk');

const client = new PulseClient('http://localhost:8081/api/v1', 'your_api_key');

client.createToken('my-room', 'user-123')
  .then(token => console.log('Token:', token))
  .catch(err => console.error(err));
` + "```" + `
`
	
	packageJSON := `{
  "name": "@pulse/sdk",
  "version": "1.0.0",
  "description": "Official Pulse Control Plane JavaScript SDK",
  "main": "index.js",
  "keywords": ["pulse", "livekit", "webrtc"],
  "author": "Pulse Team",
  "license": "MIT"
}
`
	
	return s.createZip(map[string]string{
		"pulse-sdk-js/index.js":    sdkCode,
		"pulse-sdk-js/README.md":   readmeCode,
		"pulse-sdk-js/package.json": packageJSON,
	})
}

func (s *DeveloperToolsService) generatePythonSDK() ([]byte, error) {
	sdkCode := `import requests
from typing import Dict, Any

class PulseClient:
    def __init__(self, base_url: str, api_key: str):
        self.base_url = base_url
        self.api_key = api_key
        self.session = requests.Session()
        self.session.headers.update({'X-Pulse-Key': api_key})
    
    def create_token(self, room_name: str, identity: str) -> Dict[str, Any]:
        """Create a LiveKit token for joining a room"""
        response = self.session.post(
            f"{self.base_url}/tokens/create",
            json={"room_name": room_name, "identity": identity}
        )
        response.raise_for_status()
        return response.json()
    
    # Add more methods as needed
`
	
	readmeCode := `# Pulse Python SDK

## Installation

` + "```bash" + `
pip install pulse-sdk
` + "```" + `

## Usage

` + "```python" + `
from pulse_sdk import PulseClient

client = PulseClient('http://localhost:8081/api/v1', 'your_api_key')

token = client.create_token('my-room', 'user-123')
print('Token:', token)
` + "```" + `
`
	
	setupPy := `from setuptools import setup, find_packages

setup(
    name='pulse-sdk',
    version='1.0.0',
    description='Official Pulse Control Plane Python SDK',
    packages=find_packages(),
    install_requires=['requests>=2.28.0'],
    python_requires='>=3.7',
    author='Pulse Team',
    license='MIT',
)
`
	
	return s.createZip(map[string]string{
		"pulse-sdk-python/pulse_sdk/__init__.py": sdkCode,
		"pulse-sdk-python/README.md":             readmeCode,
		"pulse-sdk-python/setup.py":              setupPy,
	})
}

func (s *DeveloperToolsService) createZip(files map[string]string) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	
	for filename, content := range files {
		f, err := w.Create(filename)
		if err != nil {
			return nil, err
		}
		_, err = io.WriteString(f, content)
		if err != nil {
			return nil, err
		}
	}
	
	err := w.Close()
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// GenerateOpenAPISpec generates OpenAPI 3.0 specification
func (s *DeveloperToolsService) GenerateOpenAPISpec() (map[string]interface{}, error) {
	// This would typically be generated from code annotations
	// For now, return a basic structure
	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "Pulse Control Plane API",
			"description": "API for managing LiveKit infrastructure",
			"version":     "1.0.0",
		},
		"servers": []map[string]string{
			{"url": "http://localhost:8081/api/v1", "description": "Development server"},
			{"url": "https://api.pulse.io/v1", "description": "Production server"},
		},
		"paths": map[string]interface{}{
			"/organizations": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "List organizations",
					"description": "Retrieve a list of all organizations",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful response",
						},
					},
				},
				"post": map[string]interface{}{
					"summary":     "Create organization",
					"description": "Create a new organization",
				},
			},
		},
	}
	
	return spec, nil
}

// RunOpenAPIGenerator generates SDKs using openapi-generator-cli
func (s *DeveloperToolsService) RunOpenAPIGenerator(specPath, language, outputDir string) error {
	// This requires openapi-generator-cli to be installed
	cmd := exec.Command("openapi-generator-cli", "generate",
		"-i", specPath,
		"-g", language,
		"-o", outputDir,
	)
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Failed to generate SDK")
		return fmt.Errorf("failed to generate SDK: %w", err)
	}
	
	return nil
}
