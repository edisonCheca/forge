package git

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/forge/forge/src/forge/core"
)

// Verificación en tiempo de compilación para asegurar que GitAdapter implementa core.GitPort.
var _ core.GitPort = (*GitAdapter)(nil)

// GitAdapter implementa la interfaz core.GitPort utilizando el comando de sistema 'git'.
type GitAdapter struct{}

// NewGitAdapter crea una nueva instancia del adaptador de Git.
func NewGitAdapter() *GitAdapter {
	return &GitAdapter{}
}

// HasStagedChanges ejecuta 'git diff --cached --quiet' para determinar si hay archivos preparados.
func (a *GitAdapter) HasStagedChanges(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", "--cached", "--quiet")
	err := cmd.Run()
	if err == nil {
		// Exit code 0 significa que no hubo diferencias (no hay cambios staged).
		return false, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if exitErr.ExitCode() == 1 {
			// Exit code 1 en 'git diff --quiet' significa que sí existen diferencias.
			return true, nil
		}
	}

	return false, fmt.Errorf("failed to check staged changes via git diff: %w", err)
}

// GetRepositoryContext obtiene el diff crudo de los archivos en el área de staging.
func (a *GitAdapter) GetRepositoryContext(ctx context.Context) (*core.RepositoryContext, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", "--cached")
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("failed to get cached diff (stderr: %s): %w", string(exitErr.Stderr), err)
		}
		return nil, fmt.Errorf("failed to get cached diff: %w", err)
	}

	return &core.RepositoryContext{
		RawDiff: string(out),
	}, nil
}

// ExecuteCommit ejecuta la confirmación formal de los cambios staged en Git.
func (a *GitAdapter) ExecuteCommit(ctx context.Context, message string) error {
	cmd := exec.CommandContext(ctx, "git", "commit", "-m", message)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute git commit (%s): %w", string(out), err)
	}
	return nil
}
