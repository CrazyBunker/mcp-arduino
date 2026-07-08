# Arduino MCP Server

MCP (Model Context Protocol) server for compiling and uploading firmware to ESP32 microcontrollers using `arduino-cli`.

## Architecture

```
├── cmd/arduino-mcp/       # Server entry point
│   └── main.go            # MCP JSON-RPC handler, tool definitions
├── internal/
│   ├── types/             # Shared data types
│   │   └── types.go       # ToolCallResult, ContentPart
│   ├── compile/           # Compilation logic
│   │   └── compile.go     # arduino-cli compile wrapper, memory parser
│   └── upload/            # Upload logic
│       └── upload.go      # USB & OTA upload, output cleaner
├── go.mod
├── README.md
├── .gitignore
├── COMPILATION_INSTRUCTION.md
└── PROMPT.md
```

## MCP Tools

All tools communicate via JSON-RPC 2.0 over stdio (standard MCP protocol).

| Tool | Description |
|---|---|
| `compile` | Compile sketch (arduino-cli compile). Returns memory usage info. |
| `upload` | Upload firmware via USB (serial port) or OTA (network). |
| `compile_and_upload` | Combined compile + upload in one step. |

### compile

**Arguments:** `sketch_path` (required), `fqbn` (optional, default: `esp32:esp32:esp32`)

```json
{
  "sketch_path": "/home/user/Arduino/MySketch",
  "fqbn": "esp32:esp32:esp32"
}
```

Returns compilation output with flash/RAM usage.

### upload

**Arguments:** `method` (usb/ota, required), `port` (required), `sketch_path` (required), `fqbn` (optional), `ota_password` (optional)

- **USB**: `arduino-cli upload -p <port> -b <fqbn> <sketch_path>`
- **OTA**: `arduino-cli upload -p <ip> -l network -b <fqbn> --upload-field password=<pass> <sketch_path>`

USB upload compiles and uploads in one command. OTA upload requires the device to be connected to WiFi with OTA enabled in the firmware.

## Requirements

- Go 1.21+
- [arduino-cli](https://arduino.github.io/arduino-cli/) installed and in PATH
- ESP32 board support: `arduino-cli core install esp32:esp32`

## Building

```bash
go build -o arduino-mcp ./cmd/arduino-mcp/
```

## Configuration for opencode

Add to `~/.config/opencode/opencode.json`:

```json
{
  "mcpServers": {
    "arduino": {
      "command": ["/path/to/arduino-mcp"],
      "type": "local",
      "enabled": true
    }
  }
}
```

## Behaviour

- **Synchronous**: All tool calls are synchronous (compilation can take 30-120s on first run). The MCP client (e.g. opencode) waits for the response.
- **Clean output**: ANSI escape codes and progress bars are stripped from output. Only relevant information is returned.
- **Memory info**: Lines containing byte counts and percentages are extracted from compilation output.
- **Error handling**: On failure, the tool returns `isError: true` with full stderr and debug info.
- **OTA password**: Default is `mySecretPass` (configurable via `ota_password` argument).
