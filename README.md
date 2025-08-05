# Fake Hue Bridge

A software-only implementation of a Philips Hue Bridge that creates fake lights compatible with diyhue and other Hue-compatible systems. Each light is displayed in its own GUI window using Fyne.

## Features

- **Multiple Light Support**: Create any number of fake lights (default: 3)
- **Individual GUI Windows**: Each light has its own window showing current color/state
- **Full Hue API Compatibility**: Compatible with both v1 and v2 (CLIP) APIs
- **diyhue Compatible**: Works with diyhue, Home Assistant, and other Hue integrations
- **SSDP Discovery**: Automatic discovery by Hue-compatible systems
- **Interactive Controls**: Click buttons and use sliders in each light window
- **Cross-Platform**: Works on Linux, Windows, and macOS

## Installation

### Prerequisites

On Linux, you'll need X11 development libraries:
```bash
sudo apt install -y libgl1-mesa-dev xorg-dev
```

### Building

```bash
git clone <this-repository>
cd fakehuebridge
go mod tidy
go build -o fakehuebridge
```

## Usage

### Basic Usage
```bash
./fakehuebridge
```
This starts 3 fake lights on port 8043.

### Custom Configuration
```bash
./fakehuebridge -lights 5 -port 8043
```

### Command Line Options
- `-lights N`: Number of fake lights to create (default: 3)
- `-port PORT`: Port for the Hue API server (default: 8043)

## GUI Windows

Each light opens in its own window containing:
- **Color Display**: Large colored rectangle showing current light state
- **ON/OFF Button**: Toggle light on/off
- **Brightness Slider**: Adjust brightness (1-254)
- **Hue Slider**: Adjust color hue (0-65535)
- **Saturation Slider**: Adjust color saturation (0-254)

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

The bridge runs with HTTPS using self-signed certificates (`server.crt` and `server.key`). When testing with curl, use the `-k` flag to ignore certificate warnings:

```bash
# Generate new certificates if needed
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes \
        -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"
```

## Integration with diyhue

1. Start the fake bridge:
   ```bash
   ./fakehuebridge -lights 10 -port 8043
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

The bridge implements SSDP (Simple Service Discovery Protocol) for automatic discovery by:
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
- Fyne v2.4.5 for GUI
- Standard Philips Hue API implementation
- SSDP/UPnP for network discovery

## License

MIT License - feel free to use and modify as needed.
