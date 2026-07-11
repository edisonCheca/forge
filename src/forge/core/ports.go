package core

import (
	"context"
)

// GitPort define el contrato para la introspección y manipulación del repositorio local.
type GitPort interface {
	HasStagedChanges(ctx context.Context) (bool, error)
	GetRepositoryContext(ctx context.Context) (*RepositoryContext, error)
	ExecuteCommit(ctx context.Context, message string) error
	GetCurrentBranch(ctx context.Context) (string, error)
	GetBranchLog(ctx context.Context, baseBranch string) ([]string, error)
	PushBranch(ctx context.Context, branch string) error
	CreatePullRequest(ctx context.Context, base, head, title, body string) (string, error)
}

// AIPort define el contrato para la comunicación con proveedores de Inteligencia Artificial.
type AIPort interface {
	GenerateCommit(ctx context.Context, req *GenerateRequest) (*CommitProposal, error)
	GeneratePullRequest(ctx context.Context, req *PRGenerateRequest) (*PRProposal, error)
}

// ConfigPort define el contrato para obtener las reglas y convenciones activas.
type ConfigPort interface {
	GetDefaultLanguage() string
	RequireConventionalCommits() bool
	GetMaxSubjectLength() int
	GetSelectedAIModel() string
	GetAIBaseURL() string
}
