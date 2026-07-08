# Arduino MCP Server

MCP (Model Context Protocol) server for ESP32 firmware development.  
Compiles Arduino sketches and uploads them via USB or OTA using `arduino-cli`.

## Tools

| Tool | Description |
|---|---|
| **compile** | Compile a sketch — returns memory usage (flash/RAM). |
| **upload** | Upload firmware via USB (serial) or OTA (Wi-Fi). |
| **compile_and_upload** | Full cycle: compile then upload in one call. |

All tools output clean LLM-readable text (no progress bars, no ANSI codes).

## Usage with opencode

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

### Examples

**Compile a sketch:**
```json
{
  "sketch_path": "/home/user/Arduino/MySketch"
}
```

**Upload via USB:**
```json
{
  "method": "usb",
  "port": "/dev/ttyUSB0",
  "sketch_path": "/home/user/Arduino/MySketch"
}
```

**Upload via OTA:**
```json
{
  "method": "ota",
  "port": "192.168.1.100",
  "sketch_path": "/home/user/Arduino/MySketch",
  "ota_password": "mySecretPass"
}
```

**Compile + upload in one step:**
```json
{
  "method": "usb",
  "port": "/dev/ttyUSB0",
  "sketch_path": "/home/user/Arduino/MySketch"
}
```

## Architecture

```
├── cmd/arduino-mcp/main.go    # Entry point, MCP JSON-RPC handler
├── internal/
│   ├── types/types.go         # Shared types (ToolCallResult)
│   ├── compile/compile.go     # arduino-cli compile wrapper
│   └── upload/upload.go       # USB & OTA upload, output sanitizer
├── go.mod
├── Makefile
└── .github/workflows/
    ├── build.yml              # CI: lint + build + artifact
    └── release.yml            # Release: multi-platform binaries
```

## Requirements

- Go 1.21+
- [arduino-cli](https://arduino.github.io/arduino-cli/) in PATH
- ESP32 core: `arduino-cli core install esp32:esp32`

## Build

```bash
go build -o arduino-mcp ./cmd/arduino-mcp/
# or
make build
```

## Behaviour Notes

- **Synchronous** — MCP client waits for the result (compilation can take 30–120s on first run).
- **OTA** — requires firmware with OTA enabled and device connected to WiFi.
- **Default OTA password** — `mySecretPass` (overridable via `ota_password`).
