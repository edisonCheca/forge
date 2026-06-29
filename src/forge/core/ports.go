package core

import (
	"context"
)

// GitPort define el contrato para la introspección y manipulación del repositorio local.
type GitPort interface {
	HasStagedChanges(ctx context.Context) (bool, error)
	GetRepositoryContext(ctx context.Context) (*RepositoryContext, error)
	ExecuteCommit(ctx context.Context, message string) error
}

// AIPort define el contrato para la comunicación con proveedores de Inteligencia Artificial.
type AIPort interface {
	GenerateCommit(ctx context.Context, req *GenerateRequest) (*CommitProposal, error)
}

// ConfigPort define el contrato para obtener las reglas y convenciones activas.
type ConfigPort interface {
	GetDefaultLanguage() string
	RequireConventionalCommits() bool
	GetMaxSubjectLength() int
	GetSelectedAIModel() string
	GetAIBaseURL() string
}
