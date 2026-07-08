package types

type ToolCallResult struct {
	Content []ContentPart `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ContentPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func ErrorResult(msg string) *ToolCallResult {
	return &ToolCallResult{
		Content: []ContentPart{
			{Type: "text", Text: msg},
		},
		IsError: true,
	}
}
