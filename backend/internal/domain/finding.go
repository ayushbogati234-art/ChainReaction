package domain

type Finding struct {
	ID                   string     `json:"id"`
	ScanID               string     `json:"scan_id"`
	PackageID            string     `json:"package_id"`
	Name                 string     `json:"name"`
	Version              string     `json:"version"`
	Ecosystem            string     `json:"ecosystem"`
	PURL                 string     `json:"purl"`
	Direct               bool       `json:"direct"`
	AdvisoryID           string     `json:"advisory_id"`
	Aliases              []string   `json:"aliases"`
	Type                 string     `json:"type"`
	DependencyConfidence Confidence `json:"dependency_confidence"`
	MatchConfidence      Confidence `json:"vulnerability_match_confidence"`
	Confidence           Confidence `json:"confidence"`
	Explanation          []string   `json:"explanation"`
	Advisories           []Advisory `json:"advisories"`
}
