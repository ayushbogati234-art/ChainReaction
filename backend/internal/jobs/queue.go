package jobs

import (
	"context"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
)

// Queue is implemented by PostgreSQL; SQL stays in transactional persistence.
type Queue interface {
	Claim(context.Context) (domain.Scan, string, error)
	Fail(context.Context, string, string, string) error
}
