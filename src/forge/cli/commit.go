package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/forge/forge/src/forge/adapters/ai"
	"github.com/forge/forge/src/forge/adapters/git"
	"github.com/forge/forge/src/forge/config"
	"github.com/forge/forge/src/forge/core"
	"github.com/spf13/cobra"
)

var issueFlag string
var commitExtraFlag string

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Auto-generate a commit message based on staged changes",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.NewAppConfig()
		apiKey := cfg.GetAIApiKey()
		baseURL := cfg.GetAIBaseURL()
		model := cfg.GetSelectedAIModel()

		reader := bufio.NewReader(os.Stdin)

		if apiKey == "" {
			fmt.Println()
			fmt.Println(Banner)
			printSeparator()
			fmt.Println(styleTitle("Asistente de Configuración Inicial"))
			fmt.Println()
			fmt.Println(styleInfo("Forge es agnóstico: puedes usar cualquier proveedor de IA (OpenRouter, Groq, Ollama, OpenAI, etc.)."))
			fmt.Println(styleInfo("Presione [Enter] en cualquier opción para aceptar el valor por defecto."))
			fmt.Println()
			printSeparator()

			fmt.Print(promptPrefix("1. URL base de la API [Por defecto: https://openrouter.ai/api/v1/chat/completions]: "))
			urlInput, err := reader.ReadString('\n')
			if err != nil && err.Error() != "EOF" {
				return fmt.Errorf("error al leer la URL base: %w", err)
			}
			baseURL = strings.TrimSpace(urlInput)
			if baseURL == "" {
				baseURL = "https://openrouter.ai/api/v1/chat/completions"
			}

			fmt.Print(promptPrefix("2. Modelo de IA [Por defecto: openai/gpt-4o-mini]: "))
			modelInput, err := reader.ReadString('\n')
			if err != nil && err.Error() != "EOF" {
				return fmt.Errorf("error al leer el modelo: %w", err)
			}
			model = strings.TrimSpace(modelInput)
			if model == "" {
				model = "openai/gpt-4o-mini"
			}

			fmt.Print(promptPrefix("3. API Key (Token de autenticación obligatorio): "))
			keyInput, err := reader.ReadString('\n')
			if err != nil && err.Error() != "EOF" {
				return fmt.Errorf("error al leer el API Key: %w", err)
			}
			apiKey = strings.TrimSpace(keyInput)
			if apiKey == "" {
				return fmt.Errorf("el API Key es obligatorio para configurar Forge")
			}

			fmt.Print(promptPrefix("4. Idioma para los commits (ej. 'es', 'en') [Por defecto: es]: "))
			langInput, err := reader.ReadString('\n')
			if err != nil && err.Error() != "EOF" {
				return fmt.Errorf("error al leer el idioma: %w", err)
			}
			language := strings.TrimSpace(langInput)
			if language == "" {
				language = "es"
			}

			if err := cfg.SaveConfig(baseURL, model, apiKey, language); err != nil {
				return fmt.Errorf("error al guardar la configuración: %w", err)
			}
			fmt.Println()
			printSeparator()
			fmt.Println(styleSuccess("Configuración guardada exitosamente en ~/.forge.json"))
			printSeparator()
			fmt.Println()
		}

		gitAdapter := git.NewGitAdapter()
		aiAdapter := ai.NewOpenAIAdapter(apiKey, baseURL, model)
		workflow := core.NewCommitWorkflow(gitAdapter, aiAdapter, cfg)

		issueRef := strings.TrimSpace(issueFlag)
		if issueRef != "" && !strings.HasPrefix(issueRef, "#") {
			issueRef = "#" + issueRef
		}

		fmt.Println(styleAction("Analizando cambios en staging..."))
		proposal, err := workflow.Execute(cmd.Context(), issueRef, commitExtraFlag)
		if err != nil {
			if errors.Is(err, core.ErrNoStagedChanges) {
				fmt.Println()
				fmt.Println(styleWarning("No hay cambios en staging. Utilice 'git add <archivos>' primero."))
				return nil
			}
			return fmt.Errorf("error al generar la propuesta de commit: %w", err)
		}

		finalSubject := proposal.Subject

		fmt.Println()
		printSeparator()
		fmt.Println(styleTitle(fmt.Sprintf("Propuesta de Commit (%s)", proposal.ModelUsed)))
		fmt.Println()
		fmt.Println(ColorWhite + finalSubject + ColorReset)
		fmt.Println()
		printSeparator()

		if issueRef == "" {
			fmt.Print(promptPrefix("¿Asociar a ID de Subtarea? [ej. 10, o Enter para omitir]: "))
			stInput, _ := reader.ReadString('\n')
			stTrimmed := strings.TrimSpace(stInput)
			if stTrimmed != "" {
				if !strings.HasPrefix(stTrimmed, "#") {
					stTrimmed = "#" + stTrimmed
				}
				lines := strings.SplitN(finalSubject, "\n", 2)
				lines[0] = strings.TrimSpace(lines[0]) + " (" + stTrimmed + ")"
				finalSubject = strings.Join(lines, "\n")
				fmt.Println(styleSuccess("Título actualizado: " + ColorWhite + lines[0] + ColorReset))
				printSeparator()
			}
		}

		fmt.Print(promptPrefix("¿Aceptar este commit? [Y/n]: "))

		input, err := reader.ReadString('\n')
		if err != nil && err.Error() != "EOF" {
			return fmt.Errorf("error al leer la confirmación del usuario: %w", err)
		}

		answer := strings.TrimSpace(input)
		if answer == "" || strings.EqualFold(answer, "y") {
			if err := gitAdapter.ExecuteCommit(cmd.Context(), finalSubject); err != nil {
				return fmt.Errorf("falló la confirmación en Git: %w", err)
			}
			fmt.Println()
			fmt.Println(styleSuccess("Commit creado exitosamente"))
		} else {
			fmt.Println()
			fmt.Println(styleInfo("Commit abortado por el usuario"))
		}

		return nil
	},
}

func init() {
	commitCmd.Flags().StringVarP(&issueFlag, "issue", "i", "", "ID de la subtarea o issue asociado (ej. 10 o #10)")
	commitCmd.Flags().StringVarP(&commitExtraFlag, "extra", "e", "", "Notas o contexto adicional sobre las decisiones y cambios realizados")
	rootCmd.AddCommand(commitCmd)
}
