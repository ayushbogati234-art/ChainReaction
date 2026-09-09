package domain

type Confidence string

const (
	ConfidenceHigh   Confidence = "HIGH"
	ConfidenceMedium Confidence = "MEDIUM"
	ConfidenceLow    Confidence = "LOW"
)

type Package struct {
	Marker, PythonVersions, Extras string
	Optional                       bool
	ID, Name, Version, PURL        string
	Direct                         bool
	Source, Groups                 string
	Confidence                     Confidence
	Limitations                    []string
}
type Edge struct {
	Constraint                                string
	Optional                                  bool
	From, To                                  string
	Direct                                    bool
	Marker, Extras, Group, Source, Confidence string
}
type Graph struct {
	InputFiles  []string
	Application Package
	Packages    []Package
	Edges       []Edge
	Incomplete  bool
	Limitations []string
}
type PathResult struct {
	Paths     [][]string `json:"paths"`
	Truncated bool       `json:"truncated"`
	Limit     int        `json:"limit"`
}
