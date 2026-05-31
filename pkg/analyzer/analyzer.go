package analyzer

// ProjectType strictly defines the frameworks DEPOLIFY supports
type ProjectType string

const (
	TypeNextJS     ProjectType = "Next.js"
	TypeReact      ProjectType = "React"
	TypeGo         ProjectType = "Go"
	TypeJavascript ProjectType = "Javascript"
	TypeExpress    ProjectType = "Express (Node.js)"
	TypePython     ProjectType = "Python"
	TypeStaticHTML ProjectType = "Static HTML/CSS"
	TypeUnknown    ProjectType = "Unknown"
)

// DeploymentDetails holds everything the orchestrator needs to deploy the app
type DeploymentDetails struct {
	Name         string      `json:"name"`
	Type         ProjectType `json:"type"`
	AbsolutePath string      `json:"absolute_path"`
	BuildCmd     string      `json:"build_command"`
	StartCmd     string      `json:"start_command"`
	DefaultPort  string      `json:"default_port"`
}

// ExtractedRepo holds the final results for the entire repository
type ExtractedRepo struct {
	Frontends []DeploymentDetails `json:"frontends"`
	Backends  []DeploymentDetails `json:"backends"`
}

// Detector is the standard function signature for the Strategy Pattern
type Detector func(dirPath string) *DeploymentDetails

// Arrays storing all our separate extraction functions
var frontendDetectors = []Detector{
	detectNextJS,
	detectReact,
	detectStaticHTML,
}

var backendDetectors = []Detector{
	detectGo,
	detectExpress,
	detectPython,
}
