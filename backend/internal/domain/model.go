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
	Paths              [][]string `json:"paths"`
	PathLengths        []int      `json:"path_lengths"`
	Applications       []string   `json:"applications"`
	Truncated          bool       `json:"truncated"`
	Limit              int        `json:"limit"`
	TotalPaths         int        `json:"total_paths"`
	CountExact         bool       `json:"count_exact"`
	CountCapped        bool       `json:"count_capped"`
	CountLimit         int        `json:"count_limit"`
	MaximumDepth       int        `json:"maximum_depth"`
	MaximumDepthExact  bool       `json:"maximum_depth_exact"`
	MaxDepthLimit      int        `json:"max_depth_limit"`
	DepthLimited       bool       `json:"depth_limited"`
	CyclesDetected     bool       `json:"cycles_detected"`
	EvidenceConfidence Confidence `json:"evidence_confidence"`
	EvidenceComplete   bool       `json:"evidence_complete"`
	Limitations        []string   `json:"limitations"`
}
