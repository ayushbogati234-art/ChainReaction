package service

import (
	"context"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/domain"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/ingest/github"
)

type ProjectStore interface {
	CreateProject(context.Context, string) (domain.Project, error)
	GetProject(context.Context, string) (domain.Project, error)
}
type ProjectService struct{ Store ProjectStore }

func (s ProjectService) CreateProject(ctx context.Context, raw string) (domain.Project, error) {
	owner, repo, err := github.ParseRepositoryURL(raw)
	if err != nil {
		return domain.Project{}, err
	}
	return s.Store.CreateProject(ctx, "https://github.com/"+owner+"/"+repo)
}
func (s ProjectService) GetProject(ctx context.Context, id string) (domain.Project, error) {
	return s.Store.GetProject(ctx, id)
}
