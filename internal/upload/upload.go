package upload

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/arduino-mcp/internal/compile"
	"github.com/arduino-mcp/internal/types"
)

type Args struct {
	Method      string `json:"method"`
	Port        string `json:"port"`
	SketchPath  string `json:"sketch_path"`
	FQBN        string `json:"fqbn"`
	OTAPassword string `json:"ota_password"`
}

func Run(argsRaw json.RawMessage) *types.ToolCallResult {
	var args Args
	if err := json.Unmarshal(argsRaw, &args); err != nil {
		return types.ErrorResult(fmt.Sprintf("Invalid arguments: %v", err))
	}

	if args.SketchPath == "" {
		return types.ErrorResult("sketch_path is required")
	}
	if args.Port == "" {
		return types.ErrorResult("port is required")
	}
	if args.FQBN == "" {
		args.FQBN = "esp32:esp32:esp32"
	}

	switch args.Method {
	case "usb":
		return uploadUSB(args)
	case "ota":
		return uploadOTA(args)
	default:
		return types.ErrorResult(fmt.Sprintf("Unknown method: %s. Use 'usb' or 'ota'.", args.Method))
	}
}

func RunCompileAndUpload(argsRaw json.RawMessage) *types.ToolCallResult {
	compileResult := compile.Run(argsRaw)
	if compileResult.IsError {
		return &types.ToolCallResult{
			Content: []types.ContentPart{
				{
					Type: "text",
					Text: "## Compile & Upload: Compilation Step Failed\n\n" + compileResult.Content[0].Text,
				},
			},
			IsError: true,
		}
	}
	return Run(argsRaw)
}

func uploadUSB(args Args) *types.ToolCallResult {
	cmd := exec.Command("arduino-cli", "upload",
		"-p", args.Port,
		"-b", args.FQBN,
		args.SketchPath,
	)

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return &types.ToolCallResult{
			Content: []types.ContentPart{
				{Type: "text", Text: formatUploadError("USB Upload Failed", args, outputStr, err)},
			},
			IsError: true,
		}
	}

	return &types.ToolCallResult{
		Content: []types.ContentPart{
			{Type: "text", Text: formatUploadSuccess("USB Upload Successful", args, outputStr)},
		},
	}
}

func uploadOTA(args Args) *types.ToolCallResult {
	password := args.OTAPassword
	if password == "" {
		password = "mySecretPass"
	}

	cmd := exec.Command("arduino-cli", "upload",
		"-p", args.Port,
		"-l", "network",
		"-b", args.FQBN,
		"--upload-field", "password="+password,
		args.SketchPath,
	)

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return &types.ToolCallResult{
			Content: []types.ContentPart{
				{Type: "text", Text: formatUploadError("OTA Upload Failed", args, outputStr, err)},
			},
			IsError: true,
		}
	}

	return &types.ToolCallResult{
		Content: []types.ContentPart{
			{Type: "text", Text: formatUploadSuccess("OTA Upload Successful", args, outputStr)},
		},
	}
}

func formatUploadSuccess(title string, args Args, output string) string {
	clean := cleanOutput(output)
	summary := parseUploadSummary(clean)
	return fmt.Sprintf("## %s\n\n**Method:** %s\n**Port:** %s\n**Sketch:** %s\n\n### Upload Result\n%s\n\n### Summary\n%s",
		title, args.Method, args.Port, args.SketchPath, summary, parseUploadMeta(clean))
}

func formatUploadError(title string, args Args, output string, err error) string {
	clean := cleanOutput(output)
	debugInfo := extractDebugInfo(clean, err)
	return fmt.Sprintf("## %s\n\n**Method:** %s\n**Port:** %s\n**Sketch:** %s\n\n**Error:** %v\n\n### Debug Info\n%s",
		title, args.Method, args.Port, args.SketchPath, err, debugInfo)
}

func cleanOutput(output string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\r`)
	cleaned := re.ReplaceAllString(output, "")
	var lines []string
	for _, line := range strings.Split(cleaned, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, "Writing at") && strings.Contains(trimmed, "[") {
			continue
		}
		if strings.Contains(trimmed, "] 100.0%") || strings.Contains(trimmed, "% ") {
			continue
		}
		lines = append(lines, trimmed)
	}
	return strings.Join(lines, "\n")
}

func parseUploadSummary(output string) string {
	var lines []string
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		lines = append(lines, trimmed)
	}
	if len(lines) == 0 {
		return "Device programmed successfully."
	}
	return strings.Join(lines, "\n")
}

func parseUploadMeta(output string) string {
	var meta []string
	for _, line := range strings.Split(output, "\n") {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "wrote") && strings.Contains(lower, "bytes") {
			meta = append(meta, line)
		}
		if strings.Contains(lower, "hash") || strings.Contains(lower, "hard reset") || strings.Contains(lower, "chip type") || strings.Contains(lower, "mac:") {
			meta = append(meta, line)
		}
	}
	if len(meta) == 0 {
		return "Firmware uploaded successfully."
	}
	return strings.Join(meta, "\n")
}

func extractDebugInfo(output string, err error) string {
	var parts []string
	if err != nil {
		parts = append(parts, fmt.Sprintf("Exit error: %v", err))
	}

	reError := regexp.MustCompile(`(?i)(error|failed|timeout|connection refused|no such file|permission denied| espota: )`)
	for _, line := range strings.Split(output, "\n") {
		if reError.MatchString(line) {
			parts = append(parts, strings.TrimSpace(line))
		}
	}

	if len(parts) == 0 {
		return "No additional debug information available."
	}
	return strings.Join(parts, "\n")
}
