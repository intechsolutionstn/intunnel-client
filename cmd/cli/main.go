package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const Version = "2.0.0"

// Server configuration
const (
	APIServer = "https://intunnel.cloud"
	Domain    = "intunnel.cloud"
)

// Embedded FRP binary (gzip compressed) - set at build time per platform
//
//go:embed frpc.gz
var embeddedFRPC []byte

// Language strings
type Lang struct {
	Welcome          string
	EnterToken       string
	InvalidToken     string
	Connecting       string
	Connected        string
	TunnelLive       string
	LocalPort        string
	PressCtrlC       string
	Disconnected     string
	Reconnecting     string
	FetchingConfig   string
	Instructions     string
	Step1            string
	Step2            string
	Step3            string
	Step4            string
	SelectLanguage   string
	Error            string
	TokenPrompt      string
	ExtractingFRP    string
	StartingTunnel   string
	EngineReady      string
	DeviceBound      string
	DeviceMismatch   string
	DeviceRegistered string
}

var languages = map[string]Lang{
	"en": {
		Welcome:          "Welcome to InTunnel",
		EnterToken:       "Enter your tunnel token",
		InvalidToken:     "Invalid token. Please check and try again.",
		Connecting:       "Connecting to InTunnel server...",
		Connected:        "Connected successfully!",
		TunnelLive:       "Your tunnel is now live at:",
		LocalPort:        "Forwarding to local port:",
		PressCtrlC:       "Press Ctrl+C to disconnect",
		Disconnected:     "Disconnected from server",
		Reconnecting:     "Reconnecting in 5 seconds...",
		FetchingConfig:   "Fetching tunnel configuration...",
		Instructions:     "How to get your token:",
		Step1:            "1. Go to https://intunnel.cloud",
		Step2:            "2. Create an account or login",
		Step3:            "3. Create a new tunnel from the dashboard",
		Step4:            "4. Copy the token from tunnel details",
		SelectLanguage:   "Select language / Choisir la langue / اختر اللغة:",
		Error:            "Error",
		TokenPrompt:      "Token: ",
		ExtractingFRP:    "Preparing tunnel engine...",
		StartingTunnel:   "Starting tunnel...",
		EngineReady:      "Tunnel engine ready!",
		DeviceBound:      "Device registered successfully!",
		DeviceMismatch:   "This token is already bound to another device.",
		DeviceRegistered: "Device:",
	},
	"fr": {
		Welcome:          "Bienvenue sur InTunnel",
		EnterToken:       "Entrez votre jeton de tunnel",
		InvalidToken:     "Jeton invalide. Veuillez vérifier et réessayer.",
		Connecting:       "Connexion au serveur InTunnel...",
		Connected:        "Connecté avec succès!",
		TunnelLive:       "Votre tunnel est maintenant actif à:",
		LocalPort:        "Redirection vers le port local:",
		PressCtrlC:       "Appuyez sur Ctrl+C pour déconnecter",
		Disconnected:     "Déconnecté du serveur",
		Reconnecting:     "Reconnexion dans 5 secondes...",
		FetchingConfig:   "Récupération de la configuration...",
		Instructions:     "Comment obtenir votre jeton:",
		Step1:            "1. Allez sur https://intunnel.cloud",
		Step2:            "2. Créez un compte ou connectez-vous",
		Step3:            "3. Créez un nouveau tunnel depuis le tableau de bord",
		Step4:            "4. Copiez le jeton depuis les détails du tunnel",
		SelectLanguage:   "Select language / Choisir la langue / اختر اللغة:",
		Error:            "Erreur",
		TokenPrompt:      "Jeton: ",
		ExtractingFRP:    "Préparation du moteur de tunnel...",
		StartingTunnel:   "Démarrage du tunnel...",
		EngineReady:      "Moteur de tunnel prêt!",
		DeviceBound:      "Appareil enregistré avec succès!",
		DeviceMismatch:   "Ce jeton est déjà lié à un autre appareil.",
		DeviceRegistered: "Appareil:",
	},
	"ar": {
		Welcome:          "مرحباً بك في InTunnel",
		EnterToken:       "أدخل رمز النفق الخاص بك",
		InvalidToken:     "رمز غير صالح. يرجى التحقق والمحاولة مرة أخرى.",
		Connecting:       "جاري الاتصال بخادم InTunnel...",
		Connected:        "تم الاتصال بنجاح!",
		TunnelLive:       "النفق الخاص بك متاح الآن على:",
		LocalPort:        "إعادة التوجيه إلى المنفذ المحلي:",
		PressCtrlC:       "اضغط Ctrl+C لقطع الاتصال",
		Disconnected:     "تم قطع الاتصال بالخادم",
		Reconnecting:     "إعادة الاتصال خلال 5 ثوانٍ...",
		FetchingConfig:   "جاري جلب إعدادات النفق...",
		Instructions:     "كيفية الحصول على الرمز:",
		Step1:            "1. اذهب إلى https://intunnel.cloud",
		Step2:            "2. أنشئ حساباً أو سجل الدخول",
		Step3:            "3. أنشئ نفقاً جديداً من لوحة التحكم",
		Step4:            "4. انسخ الرمز من تفاصيل النفق",
		SelectLanguage:   "Select language / Choisir la langue / اختر اللغة:",
		Error:            "خطأ",
		TokenPrompt:      "الرمز: ",
		ExtractingFRP:    "جاري تحضير محرك النفق...",
		StartingTunnel:   "جاري تشغيل النفق...",
		EngineReady:      "محرك النفق جاهز!",
		DeviceBound:      "تم تسجيل الجهاز بنجاح!",
		DeviceMismatch:   "هذا الرمز مرتبط بجهاز آخر.",
		DeviceRegistered: "الجهاز:",
	},
}

var currentLang Lang

// Colors
const (
	Reset  = "\033[0m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Red    = "\033[31m"
	Bold   = "\033[1m"
)

// TunnelConfig from API
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

// DeviceInfo for API request
type DeviceRequest struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func printBanner() {
	fmt.Println()
	fmt.Println(Cyan + "  ╦┌┐┌╔╦╗┬ ┬┌┐┌┌┐┌┌─┐┬  " + Reset)
	fmt.Println(Cyan + "  ║│││ ║ │ │││││││├┤ │  " + Reset)
	fmt.Println(Cyan + "  ╩┘└┘ ╩ └─┘┘└┘┘└┘└─┘┴─┘" + Reset)
	fmt.Println()
	fmt.Println(Blue + "  Secure Tunneling Platform" + Reset)
	fmt.Println(Purple + "        Version " + Version + Reset)
	fmt.Println()
}

func printStatus(status, message string) {
	var prefix string
	switch status {
	case "info":
		prefix = Blue + "[i]" + Reset
	case "success":
		prefix = Green + "[✓]" + Reset
	case "warning":
		prefix = Yellow + "[!]" + Reset
	case "error":
		prefix = Red + "[✗]" + Reset
	case "tunnel":
		prefix = Purple + "[⚡]" + Reset
	case "download":
		prefix = Cyan + "[↓]" + Reset
	case "device":
		prefix = Cyan + "[🔒]" + Reset
	}
	fmt.Printf("%s %s\n", prefix, message)
}

func selectLanguage() string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(languages["en"].SelectLanguage)
	fmt.Println()
	fmt.Println("  1. English")
	fmt.Println("  2. Français")
	fmt.Println("  3. العربية")
	fmt.Println()
	fmt.Print("Choice (1-3): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	switch input {
	case "2", "fr":
		return "fr"
	case "3", "ar":
		return "ar"
	default:
		return "en"
	}
}

func showInstructions() {
	fmt.Println()
	fmt.Println(Yellow + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + Reset)
	fmt.Println(Bold + currentLang.Instructions + Reset)
	fmt.Println()
	fmt.Println("  " + currentLang.Step1)
	fmt.Println("  " + currentLang.Step2)
	fmt.Println("  " + currentLang.Step3)
	fmt.Println("  " + currentLang.Step4)
	fmt.Println(Yellow + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + Reset)
	fmt.Println()
}

func promptToken() string {
	reader := bufio.NewReader(os.Stdin)
	showInstructions()
	fmt.Println(currentLang.EnterToken + ":")
	fmt.Println()
	fmt.Print(Cyan + currentLang.TokenPrompt + Reset)
	token, _ := reader.ReadString('\n')
	return strings.TrimSpace(token)
}

// getDeviceID generates a unique device identifier based on machine characteristics
func getDeviceID() string {
	var components []string

	// Get hostname
	hostname, err := os.Hostname()
	if err == nil {
		components = append(components, hostname)
	}

	// Get OS and architecture
	components = append(components, runtime.GOOS)
	components = append(components, runtime.GOARCH)

	// Get home directory (unique per user)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		components = append(components, homeDir)
	}

	// Get username
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME") // Windows
	}
	if username != "" {
		components = append(components, username)
	}

	// Platform-specific machine ID
	machineID := getMachineID()
	if machineID != "" {
		components = append(components, machineID)
	}

	// Create hash of all components
	data := strings.Join(components, "|")
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// getMachineID gets platform-specific machine identifier
func getMachineID() string {
	switch runtime.GOOS {
	case "linux":
		// Try to read machine-id
		if data, err := os.ReadFile("/etc/machine-id"); err == nil {
			return strings.TrimSpace(string(data))
		}
		if data, err := os.ReadFile("/var/lib/dbus/machine-id"); err == nil {
			return strings.TrimSpace(string(data))
		}
	case "darwin":
		// macOS: use IOPlatformUUID
		cmd := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "IOPlatformUUID") {
					parts := strings.Split(line, "=")
					if len(parts) == 2 {
						uuid := strings.TrimSpace(parts[1])
						uuid = strings.Trim(uuid, "\"")
						return uuid
					}
				}
			}
		}
	case "windows":
		// Windows: use MachineGuid from registry
		cmd := exec.Command("reg", "query", "HKEY_LOCAL_MACHINE\\SOFTWARE\\Microsoft\\Cryptography", "/v", "MachineGuid")
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "MachineGuid") {
					parts := strings.Fields(line)
					if len(parts) >= 3 {
						return parts[len(parts)-1]
					}
				}
			}
		}
	}
	return ""
}

// getDeviceName returns a human-readable device name
func getDeviceName() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "Unknown"
	}

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
	printStatus("info", currentLang.FetchingConfig)

	// Get device info
	deviceID := getDeviceID()
	deviceName := getDeviceName()

	printStatus("device", currentLang.DeviceRegistered+" "+deviceName)

	// Create request with device info
	url := fmt.Sprintf("%s/api/tunnel/config/%s", APIServer, token)

	reqBody := DeviceRequest{
		DeviceID:   deviceID,
		DeviceName: deviceName,
	}
	jsonBody, _ := json.Marshal(reqBody)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var config TunnelConfig
	err = json.Unmarshal(body, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	if !config.Success {
		// Check for device mismatch error
		if config.ErrorCode == "DEVICE_MISMATCH" {
			return nil, fmt.Errorf("%s\nBound to: %s", currentLang.DeviceMismatch, config.BoundDevice)
		}
		return nil, fmt.Errorf(config.Message)
	}

	return &config, nil
}

func getTempDir() string {
	return os.TempDir()
}

func getFRPPath() string {
	tempDir := getTempDir()
	if runtime.GOOS == "windows" {
		return filepath.Join(tempDir, "intunnel-frpc.exe")
	}
	return filepath.Join(tempDir, "intunnel-frpc")
}

func extractEmbeddedFRP() error {
	frpPath := getFRPPath()

	// Check if already extracted
	if _, err := os.Stat(frpPath); err == nil {
		return nil
	}

	printStatus("info", currentLang.ExtractingFRP)

	// Decompress gzip
	gzReader, err := gzip.NewReader(bytes.NewReader(embeddedFRPC))
	if err != nil {
		return fmt.Errorf("failed to decompress: %v", err)
	}
	defer gzReader.Close()

	// Write to temp file
	out, err := os.Create(frpPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, gzReader)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	// Make executable on Unix
	if runtime.GOOS != "windows" {
		os.Chmod(frpPath, 0755)
	}

	printStatus("success", currentLang.EngineReady)
	return nil
}

func createConfig(userToken string, config *TunnelConfig) (string, error) {
	tempDir := getTempDir()
	configPath := filepath.Join(tempDir, "intunnel-frpc.toml")

	// Build config with FRP auth token from server
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

	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		return "", err
	}
	return configPath, nil
}

func runFRP(configPath string, config *TunnelConfig) {
	frpPath := getFRPPath()
	fmt.Println()
	printStatus("success", currentLang.Connected)
	fmt.Println()
	fmt.Printf("%s  %s%s\n", Green, currentLang.TunnelLive, Reset)
	fmt.Printf("%s  https://%s.%s%s\n", Cyan+Bold, config.Subdomain, config.Domain, Reset)
	fmt.Println()
	fmt.Printf("%s  %s %d%s\n", Blue, currentLang.LocalPort, config.LocalPort, Reset)
	fmt.Println()
	printStatus("info", currentLang.PressCtrlC)
	fmt.Println()
	for {
		cmd := exec.Command(frpPath, "-c", configPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			printStatus("warning", currentLang.Disconnected)
			printStatus("info", currentLang.Reconnecting)
			time.Sleep(5 * time.Second)
		}
	}
}

func waitForExit() {
	fmt.Println()
	fmt.Println("Press Enter to exit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func main() {
	clearScreen()
	printBanner()
	langCode := selectLanguage()
	currentLang = languages[langCode]
	clearScreen()
	printBanner()
	fmt.Println(Bold + currentLang.Welcome + Reset)
	fmt.Println()

	token := promptToken()
	if len(token) < 10 {
		printStatus("error", currentLang.InvalidToken)
		waitForExit()
		os.Exit(1)
	}

	config, err := fetchConfig(token)
	if err != nil {
		printStatus("error", currentLang.InvalidToken)
		printStatus("error", err.Error())
		waitForExit()
		os.Exit(1)
	}

	fmt.Println()
	printStatus("success", currentLang.DeviceBound)
	printStatus("tunnel", fmt.Sprintf("Tunnel: %s", config.Name))
	printStatus("info", fmt.Sprintf("Subdomain: %s.%s", config.Subdomain, config.Domain))
	fmt.Println()

	err = extractEmbeddedFRP()
	if err != nil {
		printStatus("error", err.Error())
		waitForExit()
		os.Exit(1)
	}

	configPath, err := createConfig(token, config)
	if err != nil {
		printStatus("error", err.Error())
		waitForExit()
		os.Exit(1)
	}

	printStatus("info", currentLang.StartingTunnel)

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println()
		printStatus("info", currentLang.Disconnected)
		os.Exit(0)
	}()

	runFRP(configPath, config)
}
