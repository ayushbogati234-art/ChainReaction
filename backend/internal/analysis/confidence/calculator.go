package confidence

import "github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"

// Combine is the conservative minimum, never a safety/exploitability score.
func Combine(dependency, match domain.Confidence) domain.Confidence {
	if dependency == domain.ConfidenceLow || match == domain.ConfidenceLow || dependency == "" || match == "" {
		return domain.ConfidenceLow
	}
	if dependency == domain.ConfidenceMedium || match == domain.ConfidenceMedium {
		return domain.ConfidenceMedium
	}
	return domain.ConfidenceHigh
}
