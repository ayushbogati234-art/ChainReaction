package domain

// TopologySnapshot is a scan-scoped, storage-neutral graph used by analysis.
type TopologySnapshot struct {
	Nodes []TopologyNode
	Edges []TopologyEdge
}

type TopologyNode struct {
	ID         string
	Kind       string
	Confidence Confidence
}

type TopologyEdge struct {
	From       string
	To         string
	Confidence Confidence
}

type PathOptions struct {
	ApplicationID string
	MaxPaths      int
	MaxDepth      int
	CountLimit    int
}

type ImpactAnalysis struct {
	BlastRadius                 BlastRadius `json:"blast_radius"`
	PackageID                   string      `json:"package_id"`
	DownstreamPackages          int         `json:"downstream_packages"`
	ImmediateDependentPackages  []string    `json:"immediate_dependent_packages"`
	TransitiveDependentPackages []string    `json:"transitive_dependent_packages"`
	ApplicationsExposed         int         `json:"applications_exposed"`
	ApplicationIDs              []string    `json:"application_ids"`
	AffectedNodeIDs             []string    `json:"affected_node_ids"`
	ReachabilityExact           bool        `json:"reachability_exact"`
	EvidenceConfidence          Confidence  `json:"evidence_confidence"`
	EvidenceComplete            bool        `json:"evidence_complete"`
	Limitations                 []string    `json:"limitations"`
	Paths                       PathResult  `json:"paths"`
}

type ScoreComponent struct {
	Numerator    float64 `json:"numerator"`
	Denominator  float64 `json:"denominator"`
	Value        float64 `json:"value"`
	Weight       float64 `json:"weight"`
	Contribution float64 `json:"contribution"`
}

type BlastRadius struct {
	Score             *int                      `json:"score"`
	Classification    *string                   `json:"classification"`
	ReasonCodes       []string                  `json:"reason_codes"`
	BlockingPackage   string                    `json:"blocking_package,omitempty"`
	TotalApplications int                       `json:"total_applications"`
	TotalPackages     int                       `json:"total_packages"`
	Components        map[string]ScoreComponent `json:"components"`
}
