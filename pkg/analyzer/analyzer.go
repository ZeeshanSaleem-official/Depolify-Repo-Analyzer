package analyzer

// ProjectType defines the supported framework types.
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

// ConflictType defines the category of deployment ambiguity.
type ConflictType string

const (
	ConflictSharedRoot     ConflictType = "SharedRoot"
	ConflictPortContention ConflictType = "PortContention"
)

// Conflict represents a deployment ambiguity such as shared-root or port contention.
type Conflict struct {
	Type        ConflictType        `json:"type"`
	Directory   string              `json:"directory"`
	Services    []DeploymentDetails `json:"services"`
	Description string              `json:"description"`
}

// DeploymentDetails contains the necessary information to build and deploy a service.
type DeploymentDetails struct {
	Name         string      `json:"name"`
	Type         ProjectType `json:"type"`
	AbsolutePath string      `json:"absolute_path"`
	BuildCmd     string      `json:"build_command"`
	StartCmd     string      `json:"start_command"`
	DefaultPort  string      `json:"default_port"`
	ConflictRef  string      `json:"conflict_ref,omitempty"`
}

// ExtractedRepo contains the comprehensive deployment details for a repository.
type ExtractedRepo struct {
	Frontends []DeploymentDetails `json:"frontends"`
	Backends  []DeploymentDetails `json:"backends"`
	Conflicts []Conflict          `json:"conflicts,omitempty"`
}

// Detector defines the signature for a framework detection strategy.
type Detector func(dirPath string) *DeploymentDetails

// Detectors for various frontend and backend frameworks.
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
