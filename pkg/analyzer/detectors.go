package analyzer

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Detect NextJS
func detectNextJS(dirPath string) *DeploymentDetails {
	if !hasDependency(dirPath, "next") {
		return nil
	}
	return &DeploymentDetails{
		Type:         TypeNextJS,
		AbsolutePath: dirPath,
		BuildCmd:     "npm run build",
		StartCmd:     "npm start",
		DefaultPort:  "3000",
	}
}

// Detect React
func detectReact(dirPath string) *DeploymentDetails {
	if !hasDependency(dirPath, "react") {
		return nil
	}
	return &DeploymentDetails{
		Type:         TypeReact,
		AbsolutePath: dirPath,
		BuildCmd:     "npm run build",
		StartCmd:     "npx serve -s build",
		DefaultPort:  "3000",
	}
}

// Detect Static Html
func detectStaticHTML(dirPath string) *DeploymentDetails {
	if hasDependency(dirPath, "index.html") && !hasDependency(dirPath, "package.json") {
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

// hasDependency safely opens package.json and checks for a specific package
func hasDependency(dirPath, depName string) bool {
	pkgPath := filepath.Join(dirPath, "package.json")
	file, err := os.ReadFile(pkgPath)
	if err != nil {
		return false
	}
	var pkg struct {
		Dependencies map[string]string `json:dependencies`
	}
	err = json.Unmarshal(file, &pkg)
	if err != nil {
		return false
	}
	_, exists := pkg.Dependencies[depName]
	return exists
}

//									Backend Detectors

// Detect Express

func detectExpress(dirPath string) *DeploymentDetails {
	if !hasDependency(dirPath, "express") {
		return nil
	}
	return &DeploymentDetails{
		Type:         TypeExpress,
		AbsolutePath: dirPath,
		BuildCmd:     "npm install",
		StartCmd:     "node index.js",
		DefaultPort:  "5000",
	}
}

// Detect Go

func detectGo(dirPath string) *DeploymentDetails {
	if !hasDependency(dirPath, "go.mod") {
		return nil
	}
	return &DeploymentDetails{
		Type:         TypeGo,
		AbsolutePath: dirPath,
		BuildCmd:     "go build -o main .",
		StartCmd:     "./main",
		DefaultPort:  "8080",
	}
}

// Detect Python
func detectPython(dirPath string) *DeploymentDetails {
	if !hasFile(dirPath, "requirements.txt") {
		return nil
	}
	return &DeploymentDetails{
		Type:         TypePython,
		AbsolutePath: dirPath,
		BuildCmd:     "pip install -r requirements.txt",
		StartCmd:     "python main.py",
		DefaultPort:  "8000",
	}
}

// hasFile checks if a file exists in a directory

func hasFile(dirPath, fileName string) bool {
	info, err := os.Stat(filepath.Join(dirPath, fileName))
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
