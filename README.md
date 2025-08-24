# Huemulator

A software-only implementation of a Philips Hue Bridge that creates fake lights compatible with diyhue and other Hue-compatible systems. Each light is displayed in its own GUI window.

The initial goal of this software was to allow people who don't have Hue lights to play with [OSC2Hue](https://github.com/ilesinge/osc2hue).

## Features

- **Multiple Light Support**: Create any number of fake lights (default: 3)
- **Individual GUI Windows**: Each light has its own window showing current color/state
- **Full Hue API Compatibility**: Compatible with both v1 and v2 (CLIP) APIs
- **diyhue Compatible**: Works with diyhue, Home Assistant, and other Hue integrations
- **SSDP Discovery**: Automatic discovery by Hue-compatible systems
- **Cross-Platform**: Works on Linux, Windows, and macOS

## Installation

### Download a pre-built binary

- Download the right pre-built binary for your platform in the [releases section](https://github.com/ilesinge/huemulator/releases/latest).

### Build locally

Install the prerequisites:

- Go 1.21

You will need to install Gio UI dependencies depending on your platform:
- [Linux dependencies](https://gioui.org/doc/install/linux)
- [macOS dependencies](https://gioui.org/doc/install/macos)
- No Windows dependencies

### Building

```bash
git clone https://github.com/ilesinge/huemulator.git
cd huemulator
go mod tidy
go build -o huemulator
```

## Usage

### Basic Usage
```bash
./huemulator
```
This starts 3 fake lights on port 8043.

### Custom Configuration
```bash
./huemulator -lights 5 -port 8043
```

### Command Line Options
- `-lights N`: Number of fake lights to create (default: 3)
- `-port PORT`: Port for the Hue API server (default: 8043)

## API Endpoints

The bridge implements both Philips Hue API v1 and v2 (CLIP API) endpoints:

### V1 API (Legacy)

#### Get All Lights
```bash
curl -k "https://localhost:8043/api/testuser/lights"
```

#### Update Light State
```bash
curl -k -X PUT -H "Content-Type: application/json" \
     -d '{"on":true,"hue":25500,"sat":254,"bri":200}' \
     "https://localhost:8043/api/testuser/lights/1/state"
```

### V2 API (CLIP API)

#### Get All Lights
```bash
curl -k "https://localhost:8043/clip/v2/resource/light"
```

#### Update Light State  
```bash
curl -k -X PUT -H "Content-Type: application/json" \
     -d '{"on":{"on":true},"dimming":{"brightness":75},"color":{"xy":{"x":0.4,"y":0.5}}}' \
     "https://localhost:8043/clip/v2/resource/light/1"
```

### UPnP Description
```bash
curl -k "https://localhost:8043/description.xml"
```

## SSL/TLS Support

The bridge runs with HTTPS using self-signed certificates (`server.crt` and `server.key`). When testing with curl, use the `-k` flag to ignore certificate warnings.

## Integration with diyhue

1. Start the fake bridge:
   ```bash
   ./huemulator -lights 10 -port 8043
   ```

2. In diyhue configuration, add the bridge IP and port

3. The fake lights will appear as standard Philips Hue lights

## Light State Properties

Each light supports:
- **on**: Boolean - Light on/off state
- **bri**: Integer (1-254) - Brightness level
- **hue**: Integer (0-65535) - Color hue
- **sat**: Integer (0-254) - Color saturation
- **ct**: Integer (153-500) - Color temperature in mireds
- **colormode**: String - Current color mode ("hs" or "ct")

## Network Discovery

The bridge implements SSDP (Simple Service Discovery Protocol) and mDNS for automatic discovery by:
- diyhue bridges
- Home Assistant
- Philips Hue mobile apps
- Other Hue-compatible systems

## Example Colors

Turn light red:
```bash
curl -k -X PUT -H "Content-Type: application/json" \
     -d '{"on":true,"hue":0,"sat":254,"bri":254}' \
     "https://localhost:8043/api/testuser/lights/1/state"
```

Turn light green:
```bash
curl -k -X PUT -H "Content-Type: application/json" \
     -d '{"on":true,"hue":25500,"sat":254,"bri":254}' \
     "https://localhost:8043/api/testuser/lights/1/state"
```

Turn light blue:
```bash
curl -k -X PUT -H "Content-Type: application/json" \
     -d '{"on":true,"hue":46920,"sat":254,"bri":254}' \
     "https://localhost:8043/api/testuser/lights/1/state"
```

## Development

Built with:
- Go 1.21+
- [Gio UI](https://gioui.org/) for the GUI
- [grandcat/zeroconf](https://github.com/grandcat/zeroconf) for mDNS registering

## License

MIT License - feel free to use and modify as needed.
