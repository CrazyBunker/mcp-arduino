package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/arduino-mcp/internal/compile"
	"github.com/arduino-mcp/internal/types"
	"github.com/arduino-mcp/internal/upload"
)

type ToolInputSchema struct {
	Type       string                    `json:"type"`
	Properties map[string]InputProperty  `json:"properties"`
	Required   []string                  `json:"required,omitempty"`
}

type InputProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema ToolInputSchema `json:"inputSchema"`
}

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		handleMessage(line)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("scan error: %v", err)
	}
}

func handleMessage(line string) {
	var msg struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id,omitempty"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params,omitempty"`
	}
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		log.Printf("parse error: %v", err)
		return
	}

	switch msg.Method {
	case "initialize":
		handleInitialize(msg.ID)
	case "notifications/initialized":
		break
	case "tools/list":
		handleToolsList(msg.ID)
	case "tools/call":
		handleToolCall(msg.ID, msg.Params)
	default:
		sendError(msg.ID, -32601, fmt.Sprintf("Method not found: %s", msg.Method))
	}
}

func handleInitialize(id json.RawMessage) {
	resp := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]string{
			"name":    "arduino-mcp",
			"version": "1.0.0",
		},
	}
	sendResult(id, resp)
}

func handleToolsList(id json.RawMessage) {
	tools := []Tool{
		{
			Name:        "compile",
			Description: "Compile an Arduino sketch for ESP32. Returns build output including memory usage.",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]InputProperty{
					"sketch_path": {
						Type:        "string",
						Description: "Path to the sketch directory or .ino file",
					},
					"fqbn": {
						Type:        "string",
						Description: "Fully Qualified Board Name (default: esp32:esp32:esp32)",
					},
				},
				Required: []string{"sketch_path"},
			},
		},
		{
			Name:        "upload",
			Description: "Upload compiled firmware to ESP32 via USB or OTA. If method is usb, uploads pre-compiled binary. If method is ota, compiles and uploads.",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]InputProperty{
					"method": {
						Type:        "string",
						Description: "Upload method: 'usb' for serial or 'ota' for over-the-air",
						Enum:        []string{"usb", "ota"},
					},
					"port": {
						Type:        "string",
						Description: "Serial port (e.g. /dev/ttyUSB0) for USB or IP address for OTA",
					},
					"sketch_path": {
						Type:        "string",
						Description: "Path to the sketch directory or .ino file",
					},
					"fqbn": {
						Type:        "string",
						Description: "Fully Qualified Board Name (default: esp32:esp32:esp32)",
					},
					"ota_password": {
						Type:        "string",
						Description: "OTA password (default: mySecretPass)",
					},
				},
				Required: []string{"method", "port", "sketch_path"},
			},
		},
		{
			Name:        "compile_and_upload",
			Description: "Compile and upload firmware in one step. First compiles the sketch, then uploads via USB or OTA.",
			InputSchema: ToolInputSchema{
				Type: "object",
				Properties: map[string]InputProperty{
					"method": {
						Type:        "string",
						Description: "Upload method: 'usb' for serial or 'ota' for over-the-air",
						Enum:        []string{"usb", "ota"},
					},
					"port": {
						Type:        "string",
						Description: "Serial port (e.g. /dev/ttyUSB0) for USB or IP address for OTA",
					},
					"sketch_path": {
						Type:        "string",
						Description: "Path to the sketch directory or .ino file",
					},
					"fqbn": {
						Type:        "string",
						Description: "Fully Qualified Board Name (default: esp32:esp32:esp32)",
					},
					"ota_password": {
						Type:        "string",
						Description: "OTA password (default: mySecretPass)",
					},
				},
				Required: []string{"method", "port", "sketch_path"},
			},
		},
	}

	sendResult(id, map[string][]Tool{"tools": tools})
}

func handleToolCall(id json.RawMessage, params json.RawMessage) {
	var call struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(params, &call); err != nil {
		sendError(id, -32602, fmt.Sprintf("Invalid params: %v", err))
		return
	}

	var result *types.ToolCallResult
	switch call.Name {
	case "compile":
		result = compile.Run(call.Arguments)
	case "upload":
		result = upload.Run(call.Arguments)
	case "compile_and_upload":
		result = upload.RunCompileAndUpload(call.Arguments)
	default:
		sendError(id, -32601, fmt.Sprintf("Tool not found: %s", call.Name))
		return
	}

	sendResult(id, result)
}

func sendResult(id json.RawMessage, result interface{}) {
	resp := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	writeJSON(resp)
}

func sendError(id json.RawMessage, code int, message string) {
	resp := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	writeJSON(resp)
}

func writeJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("json marshal error: %v", err)
		return
	}
	fmt.Println(string(data))
}
