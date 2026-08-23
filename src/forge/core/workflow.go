package core

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrNoStagedChanges es el error centinela retornado cuando no existen cambios en el staging area.
var ErrNoStagedChanges = errors.New("no staged changes found")

// ExtractBranchInfo parsea el nombre de una rama para obtener el StoryID (ej. "#6") y el StoryCode (ej. "HU-2").
func ExtractBranchInfo(branch string) (string, string) {
	branchPart := branch
	if idx := strings.LastIndex(branch, "/"); idx != -1 {
		branchPart = branch[idx+1:]
	}

	var storyID string
	var storyCode string

	reStoryIDStart := regexp.MustCompile(`^([0-9]+)-(.*)$`)
	if matches := reStoryIDStart.FindStringSubmatch(branchPart); len(matches) > 2 {
		storyID = "#" + matches[1]
		branchPart = matches[2]
	}

	reCode := regexp.MustCompile(`(?i)([a-z]+-[0-9]+)`)
	if codeMatch := reCode.FindStringSubmatch(branchPart); len(codeMatch) > 1 {
		storyCode = strings.ToUpper(codeMatch[1])
	}

	return storyID, storyCode
}

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
func (w *CommitWorkflow) Execute(ctx context.Context, issueID, extraContext, verbosity string) (*CommitProposal, error) {
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
		IssueID:            issueID,
		ExtraContext:       extraContext,
		Verbosity:          verbosity,
	}

	// Paso D: Consultar al proveedor de Inteligencia Artificial
	proposal, err := w.ai.GenerateCommit(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate commit proposal via AI: %w", err)
	}

	// Paso E: Retornar la propuesta generada
	return proposal, nil
}

// PRWorkflow orquesta el flujo de negocio para la generación de Pull Requests asistidos por IA.
type PRWorkflow struct {
	git    GitPort
	ai     AIPort
	config ConfigPort
}

// NewPRWorkflow crea una nueva instancia del orquestador de Pull Requests.
func NewPRWorkflow(git GitPort, ai AIPort, config ConfigPort) *PRWorkflow {
	return &PRWorkflow{
		git:    git,
		ai:     ai,
		config: config,
	}
}

// Execute ejecuta la introspección de rama y consulta a la IA para sintetizar la descripción del PR.
func (w *PRWorkflow) Execute(ctx context.Context, branch, baseBranch string, commitLogs []string, extraContext string) (*PRProposal, error) {
	storyID, storyCode := ExtractBranchInfo(branch)
	req := &PRGenerateRequest{
		Branch:       branch,
		BaseBranch:   baseBranch,
		CommitLogs:   commitLogs,
		Language:     w.config.GetDefaultLanguage(),
		StoryID:      storyID,
		StoryCode:    storyCode,
		ExtraContext: extraContext,
	}

	proposal, err := w.ai.GeneratePullRequest(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PR proposal via AI: %w", err)
	}
	return proposal, nil
}
