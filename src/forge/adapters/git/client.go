package git

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

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

// GetCurrentBranch ejecuta 'git branch --show-current' para obtener la rama activa.
func (a *GitAdapter) GetCurrentBranch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetBranchLog obtiene los commits de la rama activa respecto a una rama base.
func (a *GitAdapter) GetBranchLog(ctx context.Context, baseBranch string) ([]string, error) {
	if baseBranch == "" {
		baseBranch = "develop"
	}
	cmd := exec.CommandContext(ctx, "git", "log", fmt.Sprintf("%s..HEAD", baseBranch), "--oneline")
	out, err := cmd.Output()
	if err != nil {
		// Si falla con la rama indicada (ej. develop no existe), intentar con main o master
		if baseBranch == "develop" {
			for _, fb := range []string{"main", "master"} {
				cmdFallback := exec.CommandContext(ctx, "git", "log", fmt.Sprintf("%s..HEAD", fb), "--oneline")
				if outFallback, errFallback := cmdFallback.Output(); errFallback == nil {
					out = outFallback
					err = nil
					break
				}
			}
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get branch log against %s: %w", baseBranch, err)
		}
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var result []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result, nil
}

// CreatePullRequest invoca a la herramienta CLI oficial de GitHub ('gh') para crear el PR.
func (a *GitAdapter) CreatePullRequest(ctx context.Context, base, head, title, body string) (string, error) {
	args := []string{"pr", "create", "--base", base, "--head", head, "--title", title, "--body", body}
	cmd := exec.CommandContext(ctx, "gh", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("gh pr create failed (%s): %w", strings.TrimSpace(string(out)), err)
	}
	return strings.TrimSpace(string(out)), nil
}

