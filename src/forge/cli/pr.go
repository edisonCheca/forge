package cli

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/forge/forge/src/forge/adapters/ai"
	"github.com/forge/forge/src/forge/adapters/git"
	"github.com/forge/forge/src/forge/config"
	"github.com/forge/forge/src/forge/core"
	"github.com/spf13/cobra"
)

var baseBranchFlag string
var extraContextFlag string

var prCmd = &cobra.Command{
	Use:     "pr",
	Aliases: []string{"pull-request"},
	Short:   "Create a standardized GitHub Pull Request from current branch",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.NewAppConfig()
		gitAdapter := git.NewGitAdapter()

		head, err := gitAdapter.GetCurrentBranch(cmd.Context())
		if err != nil || head == "" {
			return fmt.Errorf("no se pudo determinar la rama actual: %w", err)
		}

		if head == baseBranchFlag || head == "main" || head == "master" {
			fmt.Println()
			fmt.Println(styleWarning(fmt.Sprintf("No se puede crear un PR desde la rama base '%s'. Cambie a una rama de funcionalidad primero.", head)))
			return nil
		}

		fmt.Println()
		fmt.Println(styleAction(fmt.Sprintf("Analizando commits en la rama '%s'...", head)))

		commits, err := gitAdapter.GetBranchLog(cmd.Context(), baseBranchFlag)
		if err != nil {
			fmt.Println(styleWarning(fmt.Sprintf("Advertencia al leer historial contra '%s': %v", baseBranchFlag, err)))
		}

		var title, body string
		if cfg.GetAIApiKey() != "" {
			fmt.Println(styleAction("Sintetizando resumen ejecutivo del PR con Inteligencia Artificial..."))
			aiAdapter := ai.NewOpenAIAdapter(cfg.GetAIApiKey(), cfg.GetAIBaseURL(), cfg.GetSelectedAIModel())
			prWorkflow := core.NewPRWorkflow(gitAdapter, aiAdapter, cfg)
			if proposal, errAI := prWorkflow.Execute(cmd.Context(), head, baseBranchFlag, commits, extraContextFlag); errAI == nil && proposal != nil {
				title = proposal.Title
				body = proposal.Body
			}
		}

		if title == "" || body == "" {
			title, body = buildPRProposal(head, commits)
		}

		if extraContextFlag != "" {
			body += "\n\n## Notas de Diseño / Decisiones\n" + extraContextFlag
		}

		printSeparator()
		fmt.Println(styleTitle("Propuesta de Pull Request:"))
		printSeparator()
		fmt.Printf("%sTítulo:%s %s\n", ColorAccentBlue, ColorReset, ColorWhite+title+ColorReset)
		fmt.Printf("%sBase:%s   %s <- %s\n\n", ColorAccentBlue, ColorReset, baseBranchFlag, head)
		fmt.Println(ColorLightGray + "Cuerpo:" + ColorReset)
		fmt.Println(ColorWhite + body + ColorReset)
		printSeparator()

		fmt.Print(promptPrefix("¿Crear este Pull Request en GitHub? [Y/n]: "))
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil && err.Error() != "EOF" {
			return fmt.Errorf("error al leer la confirmación del usuario: %w", err)
		}

		answer := strings.TrimSpace(input)
		if answer == "" || strings.EqualFold(answer, "y") {
			fmt.Println()
			fmt.Println(styleAction(fmt.Sprintf("Sincronizando rama '%s' con origin...", head)))
			if err := gitAdapter.PushBranch(cmd.Context(), head); err != nil {
				return fmt.Errorf("falló al subir la rama al servidor remoto: %w", err)
			}
			fmt.Println(styleAction("Invocando GitHub CLI (gh pr create)..."))
			out, err := gitAdapter.CreatePullRequest(cmd.Context(), baseBranchFlag, head, title, body)
			if err != nil {
				return fmt.Errorf("error al crear el PR en GitHub: %w\nSalida: %s", err, out)
			}
			fmt.Println()
			fmt.Println(styleSuccess(fmt.Sprintf("Pull Request creado exitosamente: %s", out)))
		} else {
			fmt.Println()
			fmt.Println(styleInfo("Creación del Pull Request abortada por el usuario."))
		}

		return nil
	},
}

func init() {
	prCmd.Flags().StringVarP(&baseBranchFlag, "base", "b", "develop", "Rama base destino para el Pull Request")
	prCmd.Flags().StringVarP(&extraContextFlag, "extra", "e", "", "Notas o contexto adicional sobre las decisiones y cambios realizados")
	rootCmd.AddCommand(prCmd)
}

func buildPRProposal(head string, commits []string) (string, string) {
	branchPart := head
	if idx := strings.LastIndex(head, "/"); idx != -1 {
		branchPart = head[idx+1:]
	}

	var storyID string
	var storyCode string
	var descPart string

	reStoryIDStart := regexp.MustCompile(`^([0-9]+)-(.*)$`)
	if matches := reStoryIDStart.FindStringSubmatch(branchPart); len(matches) > 2 {
		storyID = "#" + matches[1]
		branchPart = matches[2]
	}

	reCode := regexp.MustCompile(`(?i)([a-z]+-[0-9]+)`)
	if codeMatch := reCode.FindStringSubmatch(branchPart); len(codeMatch) > 1 {
		storyCode = strings.ToUpper(codeMatch[1])
		branchPart = reCode.ReplaceAllString(branchPart, "")
	}

	descPart = strings.ReplaceAll(branchPart, "-", " ")
	descPart = strings.TrimSpace(descPart)
	words := strings.Fields(descPart)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	descFormatted := strings.Join(words, " ")
	if descFormatted == "" {
		descFormatted = "Cambios en funcionalidad"
	}

	// Tipo inferido por convención
	prefix := "feat"
	scope := "core"
	if strings.Contains(strings.ToLower(head), "fix") || strings.Contains(strings.ToLower(head), "bug") {
		prefix = "fix"
	} else if strings.Contains(strings.ToLower(head), "doc") {
		prefix = "docs"
	}

	if strings.Contains(strings.ToLower(head), "auth") || strings.Contains(strings.ToLower(head), "rbac") {
		scope = "auth"
	} else if strings.Contains(strings.ToLower(head), "api") {
		scope = "api"
	} else if strings.Contains(strings.ToLower(head), "ui") || strings.Contains(strings.ToLower(head), "cli") {
		scope = "cli"
	}

	var title string
	if storyCode != "" && storyID != "" {
		title = fmt.Sprintf("%s(%s): [%s] %s (%s)", prefix, scope, storyCode, descFormatted, storyID)
	} else if storyID != "" {
		title = fmt.Sprintf("%s(%s): %s (%s)", prefix, scope, descFormatted, storyID)
	} else if storyCode != "" {
		title = fmt.Sprintf("%s(%s): [%s] %s", prefix, scope, storyCode, descFormatted)
	} else {
		title = fmt.Sprintf("%s(%s): %s", prefix, scope, descFormatted)
	}

	// Extraer subtareas resolutivas de los commits
	reIssue := regexp.MustCompile(`#([0-9]+)`)
	var subtasks []string
	seen := make(map[string]bool)

	for _, c := range commits {
		matches := reIssue.FindAllStringSubmatch(c, -1)
		for _, m := range matches {
			if len(m) > 1 {
				issueRef := "#" + m[1]
				if issueRef != storyID && !seen[issueRef] {
					seen[issueRef] = true
					subtasks = append(subtasks, issueRef)
				}
			}
		}
	}

	var bodyBuilder strings.Builder
	bodyBuilder.WriteString("## Historia de Usuario\n")
	if storyID != "" {
		bodyBuilder.WriteString(fmt.Sprintf("Closes %s\n", storyID))
	} else {
		bodyBuilder.WriteString("Closes #[ID]\n")
	}

	if len(subtasks) > 0 {
		bodyBuilder.WriteString("\n## Subtareas Completadas\n")
		for _, st := range subtasks {
			bodyBuilder.WriteString(fmt.Sprintf("- Resolves %s\n", st))
		}
	} else if len(commits) > 0 {
		bodyBuilder.WriteString("\n## Subtareas Completadas\n")
		for _, c := range commits {
			cleanSubject := c
			if parts := strings.SplitN(c, " ", 2); len(parts) == 2 {
				cleanSubject = parts[1]
			}
			bodyBuilder.WriteString(fmt.Sprintf("- %s\n", cleanSubject))
		}
	}

	return title, strings.TrimSpace(bodyBuilder.String())
}
