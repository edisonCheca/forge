package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/forge/forge/src/forge/adapters/git"
	"github.com/forge/forge/src/forge/core"
	"github.com/spf13/cobra"
)

var (
	startTypeFlag string
	startHUFlag   string
	startBaseFlag string
)

var startCmd = &cobra.Command{
	Use:     "start [issue] [descripción]",
	Aliases: []string{"branch", "new-branch"},
	Short:   "Create and checkout a new standardized Git branch for an issue or HU",
	RunE: func(cmd *cobra.Command, args []string) error {
		gitAdapter := git.NewGitAdapter()
		reader := bufio.NewReader(os.Stdin)

		var issueID string
		var description string

		if len(args) >= 1 {
			issueID = strings.TrimSpace(args[0])
		}
		if len(args) >= 2 {
			description = strings.TrimSpace(strings.Join(args[1:], " "))
		}

		branchType := strings.TrimSpace(startTypeFlag)
		huCode := strings.TrimSpace(startHUFlag)
		baseBranch := strings.TrimSpace(startBaseFlag)

		// 1. Si se pasó el issue pero no la descripción, consultar automáticamente en GitHub
		if issueID != "" && description == "" {
			fmt.Println()
			fmt.Println(styleAction(fmt.Sprintf("Consultando detalles del Issue #%s en GitHub...", strings.TrimPrefix(issueID, "#"))))
			issueInfo, errFetch := gitAdapter.GetIssueDetails(cmd.Context(), issueID)
			if errFetch == nil && issueInfo != nil {
				inferredType, inferredHU, inferredDesc := core.ParseIssueForBranch(issueInfo)
				if !cmd.Flags().Changed("type") && inferredType != "" {
					branchType = inferredType
				}
				if !cmd.Flags().Changed("hu") && inferredHU != "" {
					huCode = inferredHU
				}
				description = inferredDesc
				fmt.Println(styleSuccess(fmt.Sprintf("Issue #%d encontrado: \"%s\"", issueInfo.Number, issueInfo.Title)))
			} else {
				fmt.Println(styleWarning(fmt.Sprintf("No se pudo obtener el issue de GitHub (%v). Proceda manualmente.", errFetch)))
			}
		}

		// 2. Si no se pasó el issueID por argumento, solicitarlo interactivamente
		if issueID == "" {
			fmt.Println()
			fmt.Println(Banner)
			printSeparator()
			fmt.Println(styleTitle("Asistente de Creación de Ramas (forge start)"))
			printSeparator()

			fmt.Print(promptPrefix("1. ID del Issue o Tarea en GitHub (ej. 5 o #5): "))
			idInput, _ := reader.ReadString('\n')
			issueID = strings.TrimSpace(idInput)
			if issueID == "" {
				return fmt.Errorf("el ID del issue o tarea es requerido para la trazabilidad")
			}

			// Intentar autocompletar tras ingresar el ID en el wizard
			fmt.Println(styleAction(fmt.Sprintf("Consultando detalles del Issue #%s en GitHub...", strings.TrimPrefix(issueID, "#"))))
			issueInfo, errFetch := gitAdapter.GetIssueDetails(cmd.Context(), issueID)
			if errFetch == nil && issueInfo != nil {
				inferredType, inferredHU, inferredDesc := core.ParseIssueForBranch(issueInfo)
				if !cmd.Flags().Changed("type") && inferredType != "" {
					branchType = inferredType
				}
				if !cmd.Flags().Changed("hu") && inferredHU != "" {
					huCode = inferredHU
				}
				description = inferredDesc
				fmt.Println(styleSuccess(fmt.Sprintf("Issue #%d encontrado: \"%s\"", issueInfo.Number, issueInfo.Title)))
			}
		}

		// 3. Fallback interactivo si tras la consulta de GitHub aún falta la descripción (ej. offline o error)
		if description == "" {
			if branchType == "" || branchType == "feat" {
				fmt.Print(promptPrefix("2. Tipo de rama [feat/fix/docs/refactor/test/chore] (Por defecto: feat): "))
				tInput, _ := reader.ReadString('\n')
				tTrimmed := strings.TrimSpace(tInput)
				if tTrimmed != "" {
					branchType = tTrimmed
				} else {
					branchType = "feat"
				}
			}

			if huCode == "" {
				fmt.Print(promptPrefix("3. Código de Historia de Usuario (ej. HU-01, o Enter para omitir): "))
				huInput, _ := reader.ReadString('\n')
				huCode = strings.TrimSpace(huInput)
			}

			fmt.Print(promptPrefix("4. Breve descripción del objetivo (ej. Control de Acceso RBAC): "))
			descInput, _ := reader.ReadString('\n')
			description = strings.TrimSpace(descInput)
			if description == "" {
				return fmt.Errorf("la descripción es requerida para construir el nombre de la rama")
			}
		}

		proposedBranch := core.BuildBranchName(branchType, issueID, huCode, description)

		currentBranch, errCurr := gitAdapter.GetCurrentBranch(cmd.Context())
		if errCurr != nil || currentBranch == "" {
			currentBranch = "actual"
		}

		baseDisplayName := currentBranch
		if baseBranch != "" {
			baseDisplayName = baseBranch
		}

		fmt.Println()
		printSeparator()
		fmt.Println(styleTitle("Propuesta de Nueva Rama:"))
		printSeparator()
		fmt.Printf("%sRama:%s %s%s%s\n", ColorAccentBlue, ColorReset, ColorWhite, proposedBranch, ColorReset)
		fmt.Printf("%sBase:%s %s\n", ColorAccentBlue, ColorReset, baseDisplayName)
		printSeparator()

		fmt.Print(promptPrefix("¿Crear y cambiar a esta rama? [Y/n]: "))
		input, err := reader.ReadString('\n')
		if err != nil && err.Error() != "EOF" {
			return fmt.Errorf("error al leer confirmación del usuario: %w", err)
		}

		answer := strings.TrimSpace(input)
		if answer == "" || strings.EqualFold(answer, "y") {
			workflow := core.NewStartWorkflow(gitAdapter)
			createdBranch, err := workflow.Execute(cmd.Context(), branchType, issueID, huCode, description, baseBranch)
			if err != nil {
				return fmt.Errorf("error al crear la rama: %w", err)
			}

			fmt.Println()
			fmt.Println(styleSuccess(fmt.Sprintf("Rama '%s' creada y activada exitosamente.", createdBranch)))
			fmt.Println(styleInfo(fmt.Sprintf("Flujo sugerido: realiza tus cambios, ejecuta 'forge commit -i %s' y al finalizar 'forge pr'.", strings.TrimPrefix(issueID, "#"))))
		} else {
			fmt.Println()
			fmt.Println(styleInfo("Creación de rama abortada por el usuario."))
		}

		return nil
	},
}

func init() {
	startCmd.Flags().StringVarP(&startTypeFlag, "type", "t", "feat", "Tipo de rama convencional (feat, fix, docs, refactor, test, chore)")
	startCmd.Flags().StringVarP(&startHUFlag, "hu", "u", "", "Código de la Historia de Usuario asociada (ej. HU-01, HU-2)")
	startCmd.Flags().StringVarP(&startBaseFlag, "base", "b", "", "Rama base origen desde la cual se bifurcará (por defecto: rama actual)")
	rootCmd.AddCommand(startCmd)
}
