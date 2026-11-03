package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/forge/forge/src/forge/adapters/ai"
	"github.com/forge/forge/src/forge/adapters/git"
	"github.com/forge/forge/src/forge/config"
	"github.com/forge/forge/src/forge/core"
	"github.com/spf13/cobra"
)

var (
	reviewStaticOnlyFlag bool
	reviewBranchFlag     bool
	reviewBaseBranchFlag string
	reviewExtraFlag      string
)

var reviewCmd = &cobra.Command{
	Use:     "review",
	Aliases: []string{"audit", "check"},
	Short:   "Audit code changes in staging or branch for security leaks, debuggers and quality",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.NewAppConfig()
		gitAdapter := git.NewGitAdapter()

		var aiAdapter core.AIPort
		if cfg.GetAIApiKey() != "" {
			aiAdapter = ai.NewOpenAIAdapter(cfg.GetAIApiKey(), cfg.GetAIBaseURL(), cfg.GetSelectedAIModel())
		}

		workflow := core.NewReviewWorkflow(gitAdapter, aiAdapter, cfg)

		targetDesc := "Área de Staging (cambios preparados)"
		if reviewBranchFlag {
			head, _ := gitAdapter.GetCurrentBranch(cmd.Context())
			targetDesc = fmt.Sprintf("Rama '%s' vs '%s'", head, reviewBaseBranchFlag)
		}

		fmt.Println()
		fmt.Println(styleAction(fmt.Sprintf("Iniciando auditoría de código sobre %s...", targetDesc)))

		result, err := workflow.Execute(cmd.Context(), reviewBranchFlag, reviewBaseBranchFlag, reviewStaticOnlyFlag, reviewExtraFlag)
		if err != nil {
			if errors.Is(err, core.ErrNoDiffToReview) {
				fmt.Println()
				fmt.Println(styleInfo("No se encontraron cambios modificados para auditar."))
				if !reviewBranchFlag {
					fmt.Println(styleInfo("Prepare sus cambios con 'git add <archivos>' o use 'forge review --branch'."))
				}
				return nil
			}
			return fmt.Errorf("error al ejecutar la auditoría: %w", err)
		}

		fmt.Println()
		printSeparator()
		fmt.Println(styleTitle("Resultados de la Auditoría (forge review)"))
		printSeparator()
		fmt.Printf("%sAlcance:%s               %s\n", ColorAccentBlue, ColorReset, targetDesc)
		fmt.Printf("%sArchivos auditados:%s    %d\n", ColorAccentBlue, ColorReset, result.FilesChecked)
		printSeparator()

		// Separar hallazgos por severidad
		var criticals []core.ReviewFinding
		var warnings []core.ReviewFinding

		for _, f := range result.Findings {
			if f.Severity == core.SeverityCritical {
				criticals = append(criticals, f)
			} else if f.Severity == core.SeverityWarning {
				warnings = append(warnings, f)
			}
		}

		// 1. Mostrar hallazgos críticos
		if len(criticals) > 0 {
			fmt.Println()
			fmt.Println(ColorError + ColorBold + "🔴 HALLAZGOS CRÍTICOS DE SEGURIDAD (ACCIÓN REQUERIDA):" + ColorReset)
			for i, c := range criticals {
				fmt.Printf("\n  %d. %s [%s]%s Archivo: %s (Línea ~%d)\n", i+1, ColorError+IconError, c.Category, ColorReset, ColorWhite+c.File+ColorReset, c.Line)
				fmt.Printf("     %sMensaje:%s %s\n", ColorLightGray, ColorReset, c.Message)
				if c.Snippet != "" {
					fmt.Printf("     %sFragmento:%s %s\n", ColorLightGray, ColorReset, ColorError+c.Snippet+ColorReset)
				}
			}
			fmt.Println()
		}

		// 2. Mostrar advertencias de calidad
		if len(warnings) > 0 {
			fmt.Println()
			fmt.Println(ColorWarning + ColorBold + "🟡 ADVERTENCIAS DE CALIDAD Y LIMPIEZA:" + ColorReset)
			for i, w := range warnings {
				fmt.Printf("\n  %d. %s [%s]%s Archivo: %s (Línea ~%d)\n", i+1, ColorWarning+IconWarning, w.Category, ColorReset, ColorWhite+w.File+ColorReset, w.Line)
				fmt.Printf("     %sMensaje:%s %s\n", ColorLightGray, ColorReset, w.Message)
				if w.Snippet != "" {
					fmt.Printf("     %sFragmento:%s %s\n", ColorLightGray, ColorReset, ColorDarkGray+w.Snippet+ColorReset)
				}
			}
			fmt.Println()
		}

		// 3. Mostrar análisis semántico de IA si existe
		if strings.TrimSpace(result.AISummary) != "" {
			fmt.Println()
			fmt.Println(styleTitle("Análisis Semántico y Sugerencias de Arquitectura (IA)"))
			fmt.Println()
			fmt.Println(ColorWhite + result.AISummary + ColorReset)
			fmt.Println()
		} else if reviewStaticOnlyFlag {
			fmt.Println()
			fmt.Println(styleInfo("Auditoría ejecutada en modo estrictamente estático (--static-only, 0 tokens)."))
		}

		printSeparator()

		// Veredicto final
		if !result.Passed {
			fmt.Println(styleError("AUDITORÍA FALLIDA: Se detectaron riesgos críticos de seguridad."))
			fmt.Println(styleWarning("Por favor elimine secretos o archivos confidenciales antes de realizar 'forge commit'."))
			return nil
		}

		if len(warnings) > 0 {
			fmt.Println(styleWarning(fmt.Sprintf("AUDITORÍA COMPLETADA CON %d ADVERTENCIA(S): Código apto para commit tras revisión manual.", len(warnings))))
		} else {
			fmt.Println(styleSuccess("AUDITORÍA EXITOSA: Código limpio, seguro y listo para producción."))
		}

		return nil
	},
}

func init() {
	reviewCmd.Flags().BoolVarP(&reviewStaticOnlyFlag, "static-only", "s", false, "Ejecutar únicamente auditoría estática determinista (sin llamadas a IA, 0 tokens)")
	reviewCmd.Flags().BoolVarP(&reviewBranchFlag, "branch", "b", false, "Auditar todos los cambios acumulados en la rama activa frente a la rama base")
	reviewCmd.Flags().StringVar(&reviewBaseBranchFlag, "base", "develop", "Rama base destino para comparar en modo rama (por defecto: develop)")
	reviewCmd.Flags().StringVarP(&reviewExtraFlag, "extra", "e", "", "Notas o contexto adicional sobre la intención del código para la IA")
	rootCmd.AddCommand(reviewCmd)
}
