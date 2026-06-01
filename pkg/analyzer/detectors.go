package analyzer

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// detectNextJS identifies Next.js applications and configures their deployment details.
func detectNextJS(dirPath string) *DeploymentDetails {
	if !hasDependency(dirPath, "next") {
		return nil
	}

	scripts := getPackageScripts(dirPath)
	rawStartScript := scripts["start"]
	smartPort := extractPort(rawStartScript, "3000")

	return &DeploymentDetails{
		Type:         TypeNextJS,
		AbsolutePath: dirPath,
		BuildCmd:     "npm run build",
		StartCmd:     "npm start",
		DefaultPort:  smartPort,
	}
}

// detectReact identifies React applications and configures their deployment details.
func detectReact(dirPath string) *DeploymentDetails {
	if !hasDependency(dirPath, "react") {
		return nil
	}

	scripts := getPackageScripts(dirPath)
	rawStartScript := scripts["start"]
	smartPort := extractPort(rawStartScript, "3000")

	// Determine the correct build directory based on Vite configuration.
	startCmd := "npx serve -s build"
	if hasFile(dirPath, "vite.config.js") || hasFile(dirPath, "vite.config.ts") {
		startCmd = "npx serve -s dist"
	}

	return &DeploymentDetails{
		Type:         TypeReact,
		AbsolutePath: dirPath,
		BuildCmd:     "npm run build",
		StartCmd:     startCmd,
		DefaultPort:  smartPort,
	}
}

// detectStaticHTML identifies static HTML applications.
func detectStaticHTML(dirPath string) *DeploymentDetails {
	if hasFile(dirPath, "index.html") && !hasFile(dirPath, "package.json") {
		return &DeploymentDetails{
			Type:         TypeStaticHTML,
			AbsolutePath: dirPath,
			BuildCmd:     "",
			StartCmd:     "npx serve",
			DefaultPort:  "80",
		}
	}
	return nil
}

// detectExpress identifies Express applications and configures their deployment details.
func detectExpress(dirPath string) *DeploymentDetails {
	if !hasDependency(dirPath, "express") {
		return nil
	}

	scripts := getPackageScripts(dirPath)
	startCmd := "node index.js" // Default fallback command
	rawStartScript := ""

	if customStart, exists := scripts["start"]; exists {
		startCmd = "npm start"
		rawStartScript = customStart
	}

	smartPort := extractPort(rawStartScript, "5000")

	return &DeploymentDetails{
		Type:         TypeExpress,
		AbsolutePath: dirPath,
		BuildCmd:     "npm install",
		StartCmd:     startCmd,
		DefaultPort:  smartPort,
	}
}

// detectGo identifies Go applications and configures their deployment details.
func detectGo(dirPath string) *DeploymentDetails {
	if !hasFile(dirPath, "go.mod") {
		return nil
	}

	// Set default build and start commands.
	buildCmd := "go build -o main ."
	startCmd := "./main"

	// Adjust build command if the standard "cmd" directory structure is used.
	if hasDir(dirPath, "cmd") {
		buildCmd = "go build -o main ./cmd/..."
	}

	// Override start command if a Procfile is present.
	if procCmd := getProcfileCommand(dirPath); procCmd != "" {
		startCmd = procCmd
	}

	return &DeploymentDetails{
		Type:         TypeGo,
		AbsolutePath: dirPath,
		BuildCmd:     buildCmd,
		StartCmd:     startCmd,
		DefaultPort:  "8080",
	}
}

// detectPython identifies Python applications and configures their deployment details.
func detectPython(dirPath string) *DeploymentDetails {
	if !hasFile(dirPath, "requirements.txt") {
		return nil
	}

	startCmd := "python main.py" // Default fallback command

	// Override start command if a Procfile is present.
	if procCmd := getProcfileCommand(dirPath); procCmd != "" {
		startCmd = procCmd
	}

	return &DeploymentDetails{
		Type:         TypePython,
		AbsolutePath: dirPath,
		BuildCmd:     "pip install -r requirements.txt",
		StartCmd:     startCmd,
		DefaultPort:  "8000",
	}
}

// hasDependency checks if a specific package is listed in package.json.
func hasDependency(dirPath, depName string) bool {
	pkgPath := filepath.Join(dirPath, "package.json")
	file, err := os.ReadFile(pkgPath)
	if err != nil {
		return false
	}
	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))
	var pkg struct {
		Dependencies map[string]string `json:"dependencies"`
	}
	err = json.Unmarshal(file, &pkg)
	if err != nil {
		return false
	}
	_, exists := pkg.Dependencies[depName]
	return exists
}

// hasFile checks if a file exists in a directory
func hasFile(dirPath, fileName string) bool {
	info, err := os.Stat(filepath.Join(dirPath, fileName))
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// hasDir checks if a directory exists inside a given path
func hasDir(dirPath, dirName string) bool {
	info, err := os.Stat(filepath.Join(dirPath, dirName))
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// getPackageScripts extracts the scripts object from package.json.
func getPackageScripts(dirPath string) map[string]string {
	pkgPath := filepath.Join(dirPath, "package.json")
	file, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil
	}

	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}

	if err := json.Unmarshal(file, &pkg); err != nil {
		return nil
	}

	return pkg.Scripts
}

// extractPort parses the start script to identify custom port declarations.
func extractPort(script string, defaultPort string) string {
	if script == "" {
		return defaultPort
	}

	// Scenario 1: Look for "PORT=8000 node server.js"
	if strings.Contains(script, "PORT=") {
		parts := strings.Split(script, "PORT=")
		fields := strings.Fields(parts[1])
		if len(fields) > 0 {
			return fields[0]
		}
	}

	// Scenario 2: Look for "next start -p 4000"
	if strings.Contains(script, "-p ") {
		parts := strings.Split(script, "-p ")
		fields := strings.Fields(parts[1])
		if len(fields) > 0 {
			return fields[0]
		}
	}

	return defaultPort
}

// getProcfileCommand extracts the web start command from a Procfile.
func getProcfileCommand(dirPath string) string {
	procfilePath := filepath.Join(dirPath, "Procfile")
	file, err := os.ReadFile(procfilePath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "web:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "web:"))
		}
	}

	return ""
}
