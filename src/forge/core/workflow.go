package core

import (
	"context"
	"errors"
	"fmt"
)

// ErrNoStagedChanges es el error centinela retornado cuando no existen cambios en el staging area.
var ErrNoStagedChanges = errors.New("no staged changes found")

// CommitWorkflow orquesta el flujo de negocio para la generación asistida de mensajes de commit.
type CommitWorkflow struct {
	git    GitPort
	ai     AIPort
	config ConfigPort
}

// NewCommitWorkflow crea una nueva instancia del orquestador inyectando sus dependencias.
func NewCommitWorkflow(git GitPort, ai AIPort, config ConfigPort) *CommitWorkflow {
	return &CommitWorkflow{
		git:    git,
		ai:     ai,
		config: config,
	}
}

// Execute ejecuta el flujo completo de introspección, construcción de contexto y consulta a la IA.
func (w *CommitWorkflow) Execute(ctx context.Context) (*CommitProposal, error) {
	// Paso A: Verificar si hay cambios preparados en staging
	hasChanges, err := w.git.HasStagedChanges(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check staged changes: %w", err)
	}
	if !hasChanges {
		return nil, ErrNoStagedChanges
	}

	// Paso B: Obtener el contexto del repositorio Git
	repoContext, err := w.git.GetRepositoryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository context: %w", err)
	}

	// Paso C: Construir la solicitud para el motor de IA combinando contexto y configuración
	req := &GenerateRequest{
		Context:            repoContext,
		Language:           w.config.GetDefaultLanguage(),
		MaxLength:          w.config.GetMaxSubjectLength(),
		ConventionalCommit: w.config.RequireConventionalCommits(),
	}

	// Paso D: Consultar al proveedor de Inteligencia Artificial
	proposal, err := w.ai.GenerateCommit(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate commit proposal via AI: %w", err)
	}

	// Paso E: Retornar la propuesta generada
	return proposal, nil
}
