# Fake Hue Light

A cross-platform fake Philips Hue light simulator written in Go. This creates virtual Hue lights that can be discovered and controlled by Hue bridges (including diyhue open source bridges).

## Features

- **Cross-platform**: Runs on Windows, macOS, and Linux
- **No dependencies**: Single executable file
- **Full color support**: RGB/HSV color control with brightness
- **Visual feedback**: Real-time web interface showing light colors
- **Auto-discovery**: Uses SSDP/UPnP for automatic bridge discovery
- **Multiple lights**: Configurable number of virtual lights
- **Hue API compatible**: Works with official and third-party bridges

## Usage

### Basic Usage
```bash
# Run with 1 light (default)
./fakehuelight

# Run with 3 lights on port 8080
./fakehuelight -lights 3 -port 8080
```

### Command Line Options
- `-lights N`: Number of fake lights to create (default: 1)
- `-port P`: HTTP port to listen on (default: 8080)

## Building

```bash
# Build for current platform
go build -o fakehuelight

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o fakehuelight.exe

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o fakehuelight-mac

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o fakehuelight-linux
```

## How it Works

1. **Discovery**: The program broadcasts SSDP responses to advertise itself as a Hue bridge
2. **API Server**: Provides HTTP endpoints compatible with Hue API
3. **Light Simulation**: Each virtual light appears as a colored rectangle in the web interface
4. **State Management**: Tracks on/off, brightness, hue, saturation for each light
5. **Web Interface**: Real-time visual feedback at http://localhost:[port+1000]

## Connecting to Bridges

### With diyhue
1. Start the fake lights: `./fakehuelight -lights 3`
2. In diyhue web interface, go to "Lights" → "Add Light"
3. The fake lights should be auto-discovered

### With Official Hue Bridge
1. Start the fake lights
2. Use the Hue app to search for new lights
3. Press the "link" button when prompted (this is simulated automatically)

## Light Controls

The web interface shows:
- **Black rectangles**: Lights that are off
- **Colored rectangles**: Current hue/saturation/brightness when on
- **Light info**: Shows the light name, ID, and on/off status
- **Real-time updates**: Colors change as you control the lights

You can control the lights through:
- Hue mobile apps
- Home automation systems
- Direct HTTP API calls
- Third-party Hue applications

## API Endpoints

The fake bridge provides these endpoints:

- `GET /api/{username}/lights` - List all lights
- `PUT /api/{username}/lights/{id}/state` - Control light state
- `GET /description.xml` - UPnP device description

## Testing

### Quick Test
Run the included test script:
```bash
# Start the application in one terminal
./fakehuelight -lights 3

# Run tests in another terminal
./test.sh
```

### Manual API Testing
```bash
# Get all lights
curl http://localhost:8080/api/testuser/lights

# Turn on light 1
curl -X PUT -H "Content-Type: application/json" -d '{"on":true}' http://localhost:8080/api/testuser/lights/1/state

# Set light 2 to red
curl -X PUT -H "Content-Type: application/json" -d '{"on":true, "hue":0, "sat":254, "bri":200}' http://localhost:8080/api/testuser/lights/2/state

# Set light 3 to blue  
curl -X PUT -H "Content-Type: application/json" -d '{"on":true, "hue":46920, "sat":254, "bri":150}' http://localhost:8080/api/testuser/lights/3/state
```

### Web Interface
Open http://localhost:9080 (or port+1000) to see real-time visual feedback of all lights.

## Troubleshooting

**Lights not discovered:**
- Check firewall settings (needs UDP 1900 and HTTP port)
- Ensure bridge and fake lights are on same network
- Try specifying a different port with `-port`

**GUI windows not showing:**
- The application uses a web interface instead of desktop windows
- Open http://localhost:[port+1000] in your browser to see the lights
- For port 8080, the web interface is at http://localhost:9080

**API not responding:**
- Check if port is already in use
- Try a different port with `-port` option

## License

MIT License - feel free to modify and distribute.
