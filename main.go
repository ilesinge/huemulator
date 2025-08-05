package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
)

// HueLight represents a single fake Hue light
type HueLight struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	State        *LightState `json:"state"`
	Type         string      `json:"type"`
	ModelID      string      `json:"modelid"`
	Manufacturer string      `json:"manufacturername"`
	SWVersion    string      `json:"swversion"`
	UniqueID     string      `json:"uniqueid"`
	mutex        sync.RWMutex
	window       fyne.Window
	colorRect    *canvas.Rectangle
	onOffButton  *widget.Button
}

// LightState represents the current state of a Hue light
type LightState struct {
	On         bool   `json:"on"`
	Brightness uint8  `json:"bri"`       // 1-254
	Hue        uint16 `json:"hue"`       // 0-65535
	Saturation uint8  `json:"sat"`       // 0-254
	ColorTemp  uint16 `json:"ct"`        // 153-500 (mireds)
	ColorMode  string `json:"colormode"` // "hs", "ct", "xy"
	Alert      string `json:"alert"`
	Effect     string `json:"effect"`
	Reachable  bool   `json:"reachable"`
}

// StateUpdate represents an update to light state
type StateUpdate struct {
	On         *bool   `json:"on,omitempty"`
	Brightness *uint8  `json:"bri,omitempty"`
	Hue        *uint16 `json:"hue,omitempty"`
	Saturation *uint8  `json:"sat,omitempty"`
	ColorTemp  *uint16 `json:"ct,omitempty"`
}

// V2 API structures for CLIP API
type V2Light struct {
	ID       string     `json:"id"`
	IDV1     string     `json:"id_v1"`
	Metadata V2Metadata `json:"metadata"`
	On       V2OnState  `json:"on"`
	Dimming  V2Dimming  `json:"dimming"`
	Color    V2Color    `json:"color,omitempty"`
	Type     string     `json:"type"`
}

type V2Metadata struct {
	Name      string `json:"name"`
	Archetype string `json:"archetype"`
}

type V2OnState struct {
	On bool `json:"on"`
}

type V2Dimming struct {
	Brightness float64 `json:"brightness"`
}

type V2Color struct {
	XY        V2XY    `json:"xy,omitempty"`
	ColorTemp V2CT    `json:"color_temperature,omitempty"`
	Gamut     V2Gamut `json:"gamut,omitempty"`
	GamutType string  `json:"gamut_type,omitempty"`
}

type V2XY struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type V2CT struct {
	Mirek int `json:"mirek"`
}

type V2Gamut struct {
	Red   V2XY `json:"red"`
	Green V2XY `json:"green"`
	Blue  V2XY `json:"blue"`
}

type V2Response struct {
	Errors []interface{} `json:"errors"`
	Data   []V2Light     `json:"data"`
}

// HueBridge represents the fake Hue Bridge
type HueBridge struct {
	lights map[string]*HueLight
	mutex  sync.RWMutex
	port   int
}

// NewHueBridge creates a new fake Hue Bridge
func NewHueBridge(port int) *HueBridge {
	return &HueBridge{
		lights: make(map[string]*HueLight),
		port:   port,
	}
}

// CreateLight creates a new light and its GUI window
func (b *HueBridge) CreateLight(id int, fyneApp fyne.App) *HueLight {
	lightID := strconv.Itoa(id)
	light := &HueLight{
		ID:           uuid.New().String(), // Generate a unique ID for the light
		Name:         fmt.Sprintf("Fake Hue Light %d", id),
		Type:         "Extended color light",
		ModelID:      "LCT016",
		Manufacturer: "Philips",
		SWVersion:    "1.65.11_r26581",
		UniqueID:     fmt.Sprintf("00:17:88:01:00:bd:ab:%02x-0b", id),
		State: &LightState{
			On:         false,
			Brightness: 254,
			Hue:        0,
			Saturation: 0,
			ColorTemp:  366,
			ColorMode:  "ct",
			Alert:      "none",
			Effect:     "none",
			Reachable:  true,
		},
	}

	// Create GUI window for this light
	light.window = fyneApp.NewWindow(fmt.Sprintf("Hue Light %d", id))
	light.window.Resize(fyne.NewSize(300, 400))

	// Position windows in a grid (note: Move() is not available in all Fyne versions)
	// x := float32((id-1) % 3 * 320)
	// y := float32((id-1) / 3 * 420)

	// Create color rectangle that fills most of the window
	light.colorRect = canvas.NewRectangle(color.RGBA{R: 50, G: 50, B: 50, A: 255})
	light.colorRect.Resize(fyne.NewSize(280, 300))

	// Create on/off button
	light.onOffButton = widget.NewButton("OFF", func() {
		light.toggleLight()
	})

	// Create brightness slider
	brightnessSlider := widget.NewSlider(1, 254)
	brightnessSlider.Value = float64(light.State.Brightness)
	brightnessSlider.OnChanged = func(value float64) {
		light.setBrightness(uint8(value))
	}

	// Create hue slider
	hueSlider := widget.NewSlider(0, 65535)
	hueSlider.Value = float64(light.State.Hue)
	hueSlider.OnChanged = func(value float64) {
		light.setHue(uint16(value))
	}

	// Create saturation slider
	satSlider := widget.NewSlider(0, 254)
	satSlider.Value = float64(light.State.Saturation)
	satSlider.OnChanged = func(value float64) {
		light.setSaturation(uint8(value))
	}

	// Create controls container
	controls := container.NewVBox(
		light.onOffButton,
		widget.NewLabel("Brightness:"),
		brightnessSlider,
		widget.NewLabel("Hue:"),
		hueSlider,
		widget.NewLabel("Saturation:"),
		satSlider,
	)

	// Main container
	content := container.NewBorder(
		nil,
		controls,
		nil,
		nil,
		light.colorRect,
	)

	light.window.SetContent(content)
	light.updateGUI()

	b.mutex.Lock()
	b.lights[lightID] = light
	b.mutex.Unlock()

	return light
}

// toggleLight toggles the light on/off
func (l *HueLight) toggleLight() {
	l.mutex.Lock()
	l.State.On = !l.State.On
	l.mutex.Unlock()
	l.updateGUI()
	log.Printf("Light %s toggled to: %v", l.ID, l.State.On)
}

// setBrightness sets the light brightness
func (l *HueLight) setBrightness(brightness uint8) {
	l.mutex.Lock()
	l.State.Brightness = brightness
	l.mutex.Unlock()
	l.updateGUI()
}

// setHue sets the light hue
func (l *HueLight) setHue(hue uint16) {
	l.mutex.Lock()
	l.State.Hue = hue
	l.State.ColorMode = "hs"
	l.mutex.Unlock()
	l.updateGUI()
}

// setSaturation sets the light saturation
func (l *HueLight) setSaturation(saturation uint8) {
	l.mutex.Lock()
	l.State.Saturation = saturation
	l.State.ColorMode = "hs"
	l.mutex.Unlock()
	l.updateGUI()
}

// updateGUI updates the visual representation of the light
func (l *HueLight) updateGUI() {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if l.State.On {
		l.onOffButton.SetText("ON")

		// Convert HSV to RGB for display
		var r, g, b uint8
		if l.State.ColorMode == "hs" {
			r, g, b = hsvToRGB(l.State.Hue, l.State.Saturation, l.State.Brightness)
		} else {
			// Color temperature mode - use warm white
			intensity := float64(l.State.Brightness) / 254.0
			r = uint8(255 * intensity)
			g = uint8(220 * intensity)
			b = uint8(180 * intensity)
		}

		l.colorRect.FillColor = color.RGBA{R: r, G: g, B: b, A: 255}
	} else {
		l.onOffButton.SetText("OFF")
		l.colorRect.FillColor = color.RGBA{R: 30, G: 30, B: 30, A: 255}
	}

	l.colorRect.Refresh()
}

// updateLightState updates light state from API call
func (l *HueLight) updateLightState(update StateUpdate) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if update.On != nil {
		l.State.On = *update.On
	}
	if update.Brightness != nil {
		l.State.Brightness = *update.Brightness
	}
	if update.Hue != nil {
		l.State.Hue = *update.Hue
		l.State.ColorMode = "hs"
	}
	if update.Saturation != nil {
		l.State.Saturation = *update.Saturation
		l.State.ColorMode = "hs"
	}
	if update.ColorTemp != nil {
		l.State.ColorTemp = *update.ColorTemp
		l.State.ColorMode = "ct"
	}

	// Update GUI in the main thread
	go func() {
		time.Sleep(10 * time.Millisecond) // Small delay to ensure thread safety
		l.updateGUI()
	}()
}

// hsvToRGB converts HSV values to RGB
func hsvToRGB(hue uint16, sat, val uint8) (r, g, b uint8) {
	h := float64(hue) / 65535.0 * 360.0
	s := float64(sat) / 254.0
	v := float64(val) / 254.0

	c := v * s
	x := c * (1 - abs(mod(h/60.0, 2)-1))
	m := v - c

	var r1, g1, b1 float64
	if h >= 0 && h < 60 {
		r1, g1, b1 = c, x, 0
	} else if h >= 60 && h < 120 {
		r1, g1, b1 = x, c, 0
	} else if h >= 120 && h < 180 {
		r1, g1, b1 = 0, c, x
	} else if h >= 180 && h < 240 {
		r1, g1, b1 = 0, x, c
	} else if h >= 240 && h < 300 {
		r1, g1, b1 = x, 0, c
	} else {
		r1, g1, b1 = c, 0, x
	}

	r = uint8((r1 + m) * 255)
	g = uint8((g1 + m) * 255)
	b = uint8((b1 + m) * 255)
	return
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func mod(x, y float64) float64 {
	return x - y*float64(int(x/y))
}

func main() {
	var numLights = flag.Int("lights", 3, "Number of fake lights to create")
	var port = flag.Int("port", 8043, "Port for the Hue API server")
	flag.Parse()

	fmt.Printf("Starting fake Hue Bridge with %d lights\n", *numLights)
	fmt.Printf("Hue API server on port %d\n", *port)

	// Create Fyne app
	fyneApp := app.New()
	// Set app metadata (if supported by Fyne version)
	// fyneApp.SetMetadata(&fyne.AppMetadata{
	//	ID:   "com.github.fakehuebridge",
	//	Name: "Fake Hue Bridge",
	// })

	// Create bridge
	bridge := NewHueBridge(*port)

	// Create lights with GUI windows
	for i := 1; i <= *numLights; i++ {
		light := bridge.CreateLight(i, fyneApp)
		light.window.Show()
	}

	// Start HTTP server for Hue API
	go startHueAPIServer(*port, bridge)

	// Start SSDP discovery service
	go startDiscoveryService(*port)

	// Run the Fyne app
	fyneApp.Run()
}

func startHueAPIServer(port int, bridge *HueBridge) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received API request: %s %s", r.Method, r.URL.Path)
		handleHueAPI(w, r, bridge)
	})
	mux.HandleFunc("/clip/v2/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received CLIP v2 API request: %s %s", r.Method, r.URL.Path)
		handleHueV2API(w, r, bridge)
	})
	mux.HandleFunc("/description.xml", handleDescription)

	log.Printf("Hue API server starting on port %d (HTTPS)", port)
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", port), "server.crt", "server.key", mux))
}

func handleHueAPI(w http.ResponseWriter, r *http.Request, bridge *HueBridge) {
	path := strings.TrimPrefix(r.URL.Path, "/api/")
	parts := strings.Split(path, "/")

	if len(parts) < 1 {
		http.Error(w, "Invalid API path", http.StatusBadRequest)
		return
	}

	// Handle different API endpoints
	if len(parts) >= 2 && parts[1] == "lights" {
		if r.Method == "GET" {
			handleGetLights(w, r, bridge)
		} else if r.Method == "PUT" && len(parts) >= 4 && parts[3] == "state" {
			handleUpdateLightState(w, r, parts[2], bridge)
		}
		return
	}

	// Default response for unknown endpoints (bridge pairing)
	response := []map[string]interface{}{
		{"success": map[string]string{"username": "fakehueuser"}},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleHueV2API(w http.ResponseWriter, r *http.Request, bridge *HueBridge) {
	path := strings.TrimPrefix(r.URL.Path, "/clip/v2/")
	parts := strings.Split(path, "/")

	if len(parts) < 1 {
		http.Error(w, "Invalid CLIP v2 API path", http.StatusBadRequest)
		return
	}

	// Handle /clip/v2/resource/light
	if len(parts) >= 2 && parts[0] == "resource" && parts[1] == "light" {
		if r.Method == "GET" {
			handleGetV2Lights(w, r, bridge)
		} else if r.Method == "PUT" && len(parts) >= 3 {
			// Handle PUT /clip/v2/resource/light/{id}
			handleUpdateV2LightState(w, r, parts[2], bridge)
		}
		return
	}

	// Default response for unknown v2 endpoints
	response := V2Response{
		Errors: []interface{}{},
		Data:   []V2Light{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetV2Lights(w http.ResponseWriter, _ *http.Request, bridge *HueBridge) {
	bridge.mutex.RLock()
	defer bridge.mutex.RUnlock()

	var v2Lights []V2Light
	for _, light := range bridge.lights {
		v2Light := convertToV2Light(light)
		v2Lights = append(v2Lights, v2Light)
	}

	response := V2Response{
		Errors: []interface{}{},
		Data:   v2Lights,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleUpdateV2LightState(w http.ResponseWriter, r *http.Request, lightID string, bridge *HueBridge) {
	var update map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Find the light
	bridge.mutex.RLock()
	light, exists := bridge.lights[lightID]
	if !exists {
		for _, l := range bridge.lights {
			if l.ID == lightID {
				light = l
				exists = true
				break
			}
		}
	}
	bridge.mutex.RUnlock()

	if !exists {
		http.Error(w, "Light not found", http.StatusNotFound)
		return
	}

	// Convert v2 format to v1 format for internal processing
	stateUpdate := convertV2ToV1StateUpdate(update)
	light.updateLightState(stateUpdate)

	// Return the updated light in v2 format
	response := V2Response{
		Errors: []interface{}{},
		Data:   []V2Light{convertToV2Light(light)},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("V2 Light %s updated via CLIP API", lightID)
}

func convertToV2Light(light *HueLight) V2Light {
	light.mutex.RLock()
	defer light.mutex.RUnlock()

	// Convert hue/sat to XY coordinates (simplified conversion)
	x, y := hueToXY(light.State.Hue, light.State.Saturation)

	v2Light := V2Light{
		ID:   light.ID,
		IDV1: "/lights/" + light.ID,
		Metadata: V2Metadata{
			Name:      light.Name,
			Archetype: "sultan_bulb",
		},
		On: V2OnState{
			On: light.State.On,
		},
		Dimming: V2Dimming{
			Brightness: float64(light.State.Brightness) / 254.0 * 100.0,
		},
		Type: "light",
	}

	// Add color information if the light supports it
	if light.State.ColorMode == "hs" {
		v2Light.Color = V2Color{
			XY: V2XY{X: x, Y: y},
			Gamut: V2Gamut{
				Red:   V2XY{X: 0.675, Y: 0.322},
				Green: V2XY{X: 0.409, Y: 0.518},
				Blue:  V2XY{X: 0.167, Y: 0.04},
			},
			GamutType: "C",
		}
	} else if light.State.ColorMode == "ct" {
		v2Light.Color = V2Color{
			ColorTemp: V2CT{
				Mirek: int(light.State.ColorTemp),
			},
		}
	}

	return v2Light
}

func convertV2ToV1StateUpdate(v2Update map[string]interface{}) StateUpdate {
	var update StateUpdate

	// Handle on/off
	if onData, exists := v2Update["on"]; exists {
		if onMap, ok := onData.(map[string]interface{}); ok {
			if on, exists := onMap["on"]; exists {
				if onBool, ok := on.(bool); ok {
					update.On = &onBool
				}
			}
		}
	}

	// Handle dimming (brightness)
	if dimmingData, exists := v2Update["dimming"]; exists {
		if dimmingMap, ok := dimmingData.(map[string]interface{}); ok {
			if brightness, exists := dimmingMap["brightness"]; exists {
				if brightnessFloat, ok := brightness.(float64); ok {
					// Convert from percentage (0-100) to Hue range (1-254)
					bri := uint8(brightnessFloat / 100.0 * 254.0)
					if bri < 1 {
						bri = 1
					}
					update.Brightness = &bri
				}
			}
		}
	}

	// Handle color
	if colorData, exists := v2Update["color"]; exists {
		if colorMap, ok := colorData.(map[string]interface{}); ok {
			// Handle XY color
			if xyData, exists := colorMap["xy"]; exists {
				if xyMap, ok := xyData.(map[string]interface{}); ok {
					if x, xExists := xyMap["x"]; xExists {
						if y, yExists := xyMap["y"]; yExists {
							if xFloat, xOk := x.(float64); xOk {
								if yFloat, yOk := y.(float64); yOk {
									// Convert XY to Hue/Sat (simplified)
									hue, sat := xyToHue(xFloat, yFloat)
									update.Hue = &hue
									update.Saturation = &sat
								}
							}
						}
					}
				}
			}

			// Handle color temperature
			if ctData, exists := colorMap["color_temperature"]; exists {
				if ctMap, ok := ctData.(map[string]interface{}); ok {
					if mirek, exists := ctMap["mirek"]; exists {
						if mirekFloat, ok := mirek.(float64); ok {
							ct := uint16(mirekFloat)
							update.ColorTemp = &ct
						}
					}
				}
			}
		}
	}

	return update
}

// Simplified XY to Hue/Sat conversion
func xyToHue(x, y float64) (uint16, uint8) {
	// This is a very simplified conversion
	// In a real implementation, you'd use proper CIE color space conversion
	hue := uint16((x * 65535.0))
	sat := uint8((y * 254.0))
	return hue, sat
}

// Simplified Hue/Sat to XY conversion
func hueToXY(hue uint16, sat uint8) (float64, float64) {
	// This is a very simplified conversion
	// In a real implementation, you'd use proper CIE color space conversion
	x := float64(hue) / 65535.0
	y := float64(sat) / 254.0
	return x, y
}

func handleGetLights(w http.ResponseWriter, _ *http.Request, bridge *HueBridge) {
	bridge.mutex.RLock()
	defer bridge.mutex.RUnlock()

	response := make(map[string]*HueLight)
	for id, light := range bridge.lights {
		response[id] = light
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleUpdateLightState(w http.ResponseWriter, r *http.Request, lightID string, bridge *HueBridge) {
	var update StateUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Find and update the light
	bridge.mutex.RLock()
	light, exists := bridge.lights[lightID]
	bridge.mutex.RUnlock()

	if !exists {
		http.Error(w, "Light not found", http.StatusNotFound)
		return
	}

	light.updateLightState(update)

	// Build response
	var responses []map[string]interface{}

	if update.On != nil {
		responses = append(responses, map[string]interface{}{
			"success": map[string]interface{}{fmt.Sprintf("/lights/%s/state/on", lightID): *update.On},
		})
	}
	if update.Brightness != nil {
		responses = append(responses, map[string]interface{}{
			"success": map[string]interface{}{fmt.Sprintf("/lights/%s/state/bri", lightID): *update.Brightness},
		})
	}
	if update.Hue != nil {
		responses = append(responses, map[string]interface{}{
			"success": map[string]interface{}{fmt.Sprintf("/lights/%s/state/hue", lightID): *update.Hue},
		})
	}
	if update.Saturation != nil {
		responses = append(responses, map[string]interface{}{
			"success": map[string]interface{}{fmt.Sprintf("/lights/%s/state/sat", lightID): *update.Saturation},
		})
	}
	if update.ColorTemp != nil {
		responses = append(responses, map[string]interface{}{
			"success": map[string]interface{}{fmt.Sprintf("/lights/%s/state/ct", lightID): *update.ColorTemp},
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)

	log.Printf("Light %s updated: on=%v, bri=%v, hue=%v, sat=%v",
		lightID, update.On, update.Brightness, update.Hue, update.Saturation)
}

func handleDescription(w http.ResponseWriter, r *http.Request) {
	description := `<?xml version="1.0" encoding="UTF-8"?>
<root xmlns="urn:schemas-upnp-org:device-1-0">
  <specVersion>
    <major>1</major>
    <minor>0</minor>
  </specVersion>
  <device>
    <deviceType>urn:schemas-upnp-org:device:Basic:1</deviceType>
    <friendlyName>Fake Hue Bridge</friendlyName>
    <manufacturer>Royal Philips Electronics</manufacturer>
    <manufacturerURL>http://www.philips.com</manufacturerURL>
    <modelDescription>Philips hue Personal Wireless Lighting</modelDescription>
    <modelName>Philips hue bridge 2012</modelName>
    <modelNumber>929000226503</modelNumber>
    <modelURL>http://www.meethue.com</modelURL>
    <serialNumber>0017880ae670</serialNumber>
    <UDN>uuid:2f402f80-da50-11e1-9b23-001788102201</UDN>
  </device>
</root>`

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(description))
}

func startDiscoveryService(port int) {
	// SSDP discovery service for Hue bridge auto-discovery
	addr, err := net.ResolveUDPAddr("udp4", "239.255.255.250:1900")
	if err != nil {
		log.Printf("Error resolving SSDP address: %v", err)
		return
	}

	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		log.Printf("Error listening for SSDP: %v", err)
		return
	}
	defer conn.Close()

	log.Println("SSDP discovery service started")

	for {
		buffer := make([]byte, 1024)
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		message := string(buffer[:n])
		if strings.Contains(message, "M-SEARCH") && strings.Contains(message, "upnp:rootdevice") {
			go handleSSDPRequest(clientAddr, port)
		}
	}
}

func handleSSDPRequest(clientAddr *net.UDPAddr, port int) {
	// Get local IP address
	localIP, err := getLocalIP()
	if err != nil {
		return
	}

	response := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
		"CACHE-CONTROL: max-age=100\r\n"+
		"EXT:\r\n"+
		"LOCATION: http://%s:%d/description.xml\r\n"+
		"SERVER: Linux/3.14.0 UPnP/1.0 IpBridge/1.65.0\r\n"+
		"ST: upnp:rootdevice\r\n"+
		"USN: uuid:2f402f80-da50-11e1-9b23-001788102201::upnp:rootdevice\r\n\r\n",
		localIP, port)

	conn, err := net.Dial("udp", clientAddr.String())
	if err != nil {
		return
	}
	defer conn.Close()

	conn.Write([]byte(response))
}

func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
