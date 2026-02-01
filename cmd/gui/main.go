package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const Version = "2.0.0"
const APIServer = "https://intunnel.cloud"

//go:embed frpc.gz
var embeddedFRPC []byte

//go:embed index.html
var indexHTML string

type TunnelConfig struct {
	Success     bool   `json:"success"`
	Subdomain   string `json:"subdomain"`
	LocalPort   int    `json:"local_port"`
	Name        string `json:"name"`
	Username    string `json:"username"`
	Message     string `json:"message"`
	ServerIP    string `json:"server_ip"`
	ServerPort  int    `json:"server_port"`
	Domain      string `json:"domain"`
	AuthToken   string `json:"auth_token"`
	ErrorCode   string `json:"error_code"`
	BoundDevice string `json:"bound_device"`
}

type DeviceRequest struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

type AppState struct {
	sync.Mutex
	Connected   bool
	TunnelName  string
	PublicURL   string
	LocalPort   int
	DeviceName  string
	Status      string
	tunnelCmd   *exec.Cmd
}

var state = &AppState{
	Status:     "ready",
	DeviceName: getDeviceName(),
}

func getDeviceID() string {
	var components []string
	hostname, _ := os.Hostname()
	components = append(components, hostname, runtime.GOOS, runtime.GOARCH)
	homeDir, _ := os.UserHomeDir()
	components = append(components, homeDir)
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	components = append(components, username)
	machineID := getMachineID()
	if machineID != "" {
		components = append(components, machineID)
	}
	data := strings.Join(components, "|")
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func getMachineID() string {
	switch runtime.GOOS {
	case "linux":
		if data, err := os.ReadFile("/etc/machine-id"); err == nil {
			return strings.TrimSpace(string(data))
		}
	case "darwin":
		cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
		output, _ := cmd.Output()
		for _, line := range strings.Split(string(output), "\n") {
			if strings.Contains(line, "IOPlatformUUID") {
				parts := strings.Split(line, "=")
				if len(parts) == 2 {
					return strings.Trim(strings.TrimSpace(parts[1]), "\"")
				}
			}
		}
	case "windows":
		cmd := exec.Command("reg", "query", "HKEY_LOCAL_MACHINE\\SOFTWARE\\Microsoft\\Cryptography", "/v", "MachineGuid")
		output, _ := cmd.Output()
		for _, line := range strings.Split(string(output), "\n") {
			if strings.Contains(line, "MachineGuid") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					return parts[len(parts)-1]
				}
			}
		}
	}
	return ""
}

func getDeviceName() string {
	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	if username == "" {
		username = "user"
	}
	osName := runtime.GOOS
	switch osName {
	case "darwin":
		osName = "macOS"
	case "windows":
		osName = "Windows"
	case "linux":
		osName = "Linux"
	}
	return fmt.Sprintf("%s@%s (%s)", username, hostname, osName)
}

func fetchConfig(token string) (*TunnelConfig, error) {
	url := fmt.Sprintf("%s/api/tunnel/config/%s", APIServer, token)
	reqBody := DeviceRequest{DeviceID: getDeviceID(), DeviceName: getDeviceName()}
	jsonBody, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Connection failed")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var config TunnelConfig
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, fmt.Errorf("Invalid server response")
	}

	if !config.Success {
		if config.ErrorCode == "DEVICE_MISMATCH" {
			return nil, fmt.Errorf("Token bound to: %s", config.BoundDevice)
		}
		return nil, fmt.Errorf(config.Message)
	}
	return &config, nil
}

func extractFRP() (string, error) {
	tempDir := os.TempDir()
	frpPath := filepath.Join(tempDir, "intunnel-frpc-gui")
	if runtime.GOOS == "windows" {
		frpPath += ".exe"
	}
	if _, err := os.Stat(frpPath); err == nil {
		return frpPath, nil
	}
	gzReader, err := gzip.NewReader(bytes.NewReader(embeddedFRPC))
	if err != nil {
		return "", err
	}
	defer gzReader.Close()
	out, err := os.Create(frpPath)
	if err != nil {
		return "", err
	}
	defer out.Close()
	io.Copy(out, gzReader)
	if runtime.GOOS != "windows" {
		os.Chmod(frpPath, 0755)
	}
	return frpPath, nil
}

func createFRPConfig(config *TunnelConfig) (string, error) {
	tempDir := os.TempDir()
	configPath := filepath.Join(tempDir, "intunnel-frpc-gui.toml")
	content := fmt.Sprintf(`serverAddr = "%s"
serverPort = %d
auth.method = "token"
auth.token = "%s"

[[proxies]]
name = "%s"
type = "http"
localPort = %d
subdomain = "%s"
`, config.ServerIP, config.ServerPort, config.AuthToken, config.Subdomain, config.LocalPort, config.Subdomain)
	return configPath, os.WriteFile(configPath, []byte(content), 0644)
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	if len(token) < 10 {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid token"})
		return
	}

	config, err := fetchConfig(token)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	frpPath, err := extractFRP()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to prepare tunnel"})
		return
	}

	configPath, err := createFRPConfig(config)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to create config"})
		return
	}

	state.Lock()
	state.tunnelCmd = exec.Command(frpPath, "-c", configPath)
	if err := state.tunnelCmd.Start(); err != nil {
		state.Unlock()
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to start tunnel"})
		return
	}

	state.Connected = true
	state.TunnelName = config.Name
	state.PublicURL = fmt.Sprintf("https://%s.%s", config.Subdomain, config.Domain)
	state.LocalPort = config.LocalPort
	state.Status = "connected"
	state.Unlock()

	go func() {
		state.tunnelCmd.Wait()
		state.Lock()
		state.Connected = false
		state.Status = "disconnected"
		state.Unlock()
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"tunnelName": config.Name,
		"publicURL":  state.PublicURL,
		"localPort":  config.LocalPort,
	})
}

func handleDisconnect(w http.ResponseWriter, r *http.Request) {
	state.Lock()
	defer state.Unlock()
	if state.tunnelCmd != nil && state.tunnelCmd.Process != nil {
		state.tunnelCmd.Process.Kill()
	}
	state.Connected = false
	state.Status = "ready"
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	state.Lock()
	defer state.Unlock()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"connected":  state.Connected,
		"tunnelName": state.TunnelName,
		"publicURL":  state.PublicURL,
		"localPort":  state.LocalPort,
		"deviceName": state.DeviceName,
		"status":     state.Status,
	})
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("index").Parse(indexHTML))
	tmpl.Execute(w, map[string]string{"Version": Version, "DeviceName": getDeviceName()})
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}

func main() {
	// Find available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Println("Failed to start server:", err)
		os.Exit(1)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Setup routes
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/connect", handleConnect)
	http.HandleFunc("/disconnect", handleDisconnect)
	http.HandleFunc("/status", handleStatus)

	url := fmt.Sprintf("http://127.0.0.1:%d", port)
	fmt.Printf("InTunnel v%s\n", Version)
	fmt.Printf("Opening %s in your browser...\n", url)
	fmt.Println("Keep this window open while using the tunnel.")
	fmt.Println("Press Ctrl+C to exit.")

	go func() {
		time.Sleep(500 * time.Millisecond)
		openBrowser(url)
	}()

	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), nil)
}
