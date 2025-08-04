package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// HueLight represents a single fake Hue light
type HueLight struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	State        *LightState `json:"state"`
	Type         string     `json:"type"`
	ModelID      string     `json:"modelid"`
	Manufacturer string     `json:"manufacturername"`
	SWVersion    string     `json:"swversion"`
	UniqueID     string     `json:"uniqueid"`
	mutex        sync.RWMutex
}

// LightState represents the current state of a Hue light
type LightState struct {
	On         bool   `json:"on"`
	Brightness uint8  `json:"bri"`      // 1-254
	Hue        uint16 `json:"hue"`      // 0-65535
	Saturation uint8  `json:"sat"`      // 0-254
	ColorTemp  uint16 `json:"ct"`       // 153-500 (mireds)
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

var lights []*HueLight
var lightsMutex sync.RWMutex
var webPort int

func main() {
	numLights := flag.Int("lights", 1, "Number of fake lights to create")
	port := flag.Int("port", 8080, "Port to listen on")
	webPort = *port + 1000
	
	flag.Parse()

	fmt.Printf("Starting %d fake Hue lights\n", *numLights)
	fmt.Printf("Hue API server on port %d\n", *port)
	fmt.Printf("Web interface at http://localhost:%d\n", webPort)

	// Initialize lights
	for i := 0; i < *numLights; i++ {
		light := createLight(i + 1)
		lights = append(lights, light)
	}

	// Start web interface server
	go startWebServer()

	// Start HTTP server for Hue API
	startHueAPIServer(*port)
}

func createLight(id int) *HueLight {
	light := &HueLight{
		ID:           strconv.Itoa(id),
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
	return light
}

func (light *HueLight) updateState(update StateUpdate) {
	light.mutex.Lock()
	defer light.mutex.Unlock()
	
	if update.On != nil {
		light.State.On = *update.On
	}
	if update.Brightness != nil {
		light.State.Brightness = *update.Brightness
	}
	if update.Hue != nil {
		light.State.Hue = *update.Hue
		light.State.ColorMode = "hs"
	}
	if update.Saturation != nil {
		light.State.Saturation = *update.Saturation
		light.State.ColorMode = "hs"
	}
	if update.ColorTemp != nil {
		light.State.ColorTemp = *update.ColorTemp
		light.State.ColorMode = "ct"
	}
}

func startWebServer() {
	http.HandleFunc("/", handleWebInterface)
	http.HandleFunc("/lights.json", handleLightsJSON)
	
	log.Printf("Web interface starting on port %d", webPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", webPort), nil))
}

func handleWebInterface(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
	<title>Fake Hue Lights</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; background: #1a1a1a; color: white; }
		.lights-container { display: flex; flex-wrap: wrap; gap: 20px; }
		.light { 
			width: 200px; 
			height: 200px; 
			border-radius: 10px;
			border: 3px solid #333;
			display: flex;
			align-items: center;
			justify-content: center;
			text-align: center;
			font-weight: bold;
			text-shadow: 1px 1px 2px rgba(0,0,0,0.7);
		}
		.light-info {
			background: rgba(0,0,0,0.3);
			padding: 10px;
			border-radius: 5px;
		}
		h1 { color: #fff; }
		.status { margin: 10px 0; font-size: 14px; }
	</style>
	<script>
		function updateLights() {
			fetch('/lights.json')
				.then(response => response.json())
				.then(lights => {
					const container = document.getElementById('lights');
					container.innerHTML = '';
					
					for (const lightId in lights) {
						const light = lights[lightId];
						const div = document.createElement('div');
						div.className = 'light';
						
						let bgColor = 'rgb(0,0,0)';
						if (light.state.on) {
							const h = light.state.hue / 65535 * 360;
							const s = light.state.sat / 254 * 100;
							const v = light.state.bri / 254 * 100;
							bgColor = 'hsl(' + h + ',' + s + '%,' + v + '%)';
						}
						
						div.style.backgroundColor = bgColor;
						div.innerHTML = '<div class="light-info"><div>' + light.name + '</div><div>ID: ' + light.id + '</div><div>' + (light.state.on ? 'ON' : 'OFF') + '</div></div>';
						container.appendChild(div);
					}
				});
		}
		
		setInterval(updateLights, 1000);
		window.onload = updateLights;
	</script>
</head>
<body>
	<h1>Fake Hue Lights Monitor</h1>
	<div class="status">
		<p>This page shows the current state of your fake Hue lights.</p>
		<p>Control them using the Hue API or any Hue-compatible app.</p>
	</div>
	<div id="lights" class="lights-container">
	</div>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleLightsJSON(w http.ResponseWriter, r *http.Request) {
	lightsMutex.RLock()
	defer lightsMutex.RUnlock()
	
	response := make(map[string]*HueLight)
	for _, light := range lights {
		response[light.ID] = light
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func startHueAPIServer(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/", handleHueAPI)
	mux.HandleFunc("/description.xml", handleDescription)
	
	log.Printf("Hue API server starting on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}

func handleHueAPI(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/")
	parts := strings.Split(path, "/")
	
	if len(parts) < 1 {
		http.Error(w, "Invalid API path", http.StatusBadRequest)
		return
	}
	
	// Handle different API endpoints
	if len(parts) >= 2 && parts[1] == "lights" {
		if r.Method == "GET" {
			handleGetLights(w, r)
		} else if r.Method == "PUT" && len(parts) >= 4 && parts[3] == "state" {
			handleUpdateLightState(w, r, parts[2])
		}
		return
	}
	
	// Default response for unknown endpoints
	response := []map[string]interface{}{
		{"success": map[string]string{"username": "fakehueuser"}},
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetLights(w http.ResponseWriter, r *http.Request) {
	lightsMutex.RLock()
	defer lightsMutex.RUnlock()
	
	response := make(map[string]*HueLight)
	for _, light := range lights {
		response[light.ID] = light
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleUpdateLightState(w http.ResponseWriter, r *http.Request, lightID string) {
	var update StateUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	lightsMutex.RLock()
	var targetLight *HueLight
	for _, light := range lights {
		if light.ID == lightID {
			targetLight = light
			break
		}
	}
	lightsMutex.RUnlock()
	
	if targetLight == nil {
		http.Error(w, "Light not found", http.StatusNotFound)
		return
	}
	
	targetLight.updateState(update)
	
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
