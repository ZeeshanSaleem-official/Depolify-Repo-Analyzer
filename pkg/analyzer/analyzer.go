package analyzer

// ProjectType strictly defines the frameworks DEPOLIFY supports
type ProjectType string

const (
	TypeNextJS     ProjectType = "Next.js"
	TypeReact      ProjectType = "React"
	TypeGo         ProjectType = "Go"
	TypeJavascript ProjectType = "Javascript"
	TypeUnknown    ProjectType = "Unkown"
)

// DeploymentDetails holds everything the orchestrator needs to deploy the app
type DeploymentDetails struct {
	Type         ProjectType `json: "type"`
	AbsolutePath string      `json:"absolute_path"`
	BuildCmd     string      `json:"build_cmd"`
	StartCmd     string      `json:"start_command"`
	DefaultPort  string      `json:"default_port"`
}

// ExtractedRepo holds the final results for the entire repository

type ExtractedRepo struct {
	Frontends []DeploymentDetails `json:"frontends"`
	Backends  []DeploymentDetails `json:"backends"`
}

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
