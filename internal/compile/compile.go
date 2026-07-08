package compile

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/arduino-mcp/internal/types"
)

type Args struct {
	SketchPath string `json:"sketch_path"`
	FQBN       string `json:"fqbn"`
}

func Run(argsRaw json.RawMessage) *types.ToolCallResult {
	var args Args
	if err := json.Unmarshal(argsRaw, &args); err != nil {
		return types.ErrorResult(fmt.Sprintf("Invalid arguments: %v", err))
	}

	if args.SketchPath == "" {
		return types.ErrorResult("sketch_path is required")
	}
	if args.FQBN == "" {
		args.FQBN = "esp32:esp32:esp32"
	}

	cmd := exec.Command("arduino-cli", "compile", "-b", args.FQBN, args.SketchPath)
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	memInfo := parseMemoryInfo(outputStr)

	if err != nil {
		return &types.ToolCallResult{
			Content: []types.ContentPart{
				{Type: "text", Text: fmt.Sprintf("## Compilation Failed\n\n**FQBN:** %s\n**Sketch:** %s\n\n```\n%s\n```\n\n### Memory Info\n%s", args.FQBN, args.SketchPath, outputStr, memInfo)},
			},
			IsError: true,
		}
	}

	return &types.ToolCallResult{
		Content: []types.ContentPart{
			{Type: "text", Text: fmt.Sprintf("## Compilation Successful\n\n**FQBN:** %s\n**Sketch:** %s\n\n```\n%s\n```\n\n### Memory Usage\n%s", args.FQBN, args.SketchPath, outputStr, memInfo)},
		},
	}
}

func parseMemoryInfo(output string) string {
	var lines []string
	re := regexp.MustCompile(`(?i)\d+\s*(?:байт|bytes|\([\d.]+%\))`)

	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && re.MatchString(trimmed) {
			lines = append(lines, trimmed)
		}
	}
	if len(lines) == 0 {
		return "No memory information found in output."
	}
	return strings.Join(lines, "\n")
}
