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

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Auto-generate a commit message based on staged changes",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Inyección de dependencias y asistente de configuración (Wizard)
		cfg := config.NewAppConfig()
		apiKey := cfg.GetAIApiKey()
		baseURL := cfg.GetAIBaseURL()
		model := cfg.GetSelectedAIModel()

		reader := bufio.NewReader(os.Stdin)

		if apiKey == "" {
			fmt.Println("===================================================================")
			fmt.Println("Forge - Asistente de Configuración Inicial")
			fmt.Println("===================================================================")
			fmt.Println("Para generar propuestas de commit, Forge requiere conectarse a una API")
			fmt.Println("compatible con OpenAI (ej. OpenRouter, OpenAI, Groq, Ollama local).")
			fmt.Println("\nPresione [Enter] en cualquier opción para aceptar el valor por defecto.")
			fmt.Println("-------------------------------------------------------------------")

			fmt.Print("1. URL base de la API [Por defecto: https://openrouter.ai/api/v1/chat/completions]: ")
			urlInput, err := reader.ReadString('\n')
			if err != nil && err.Error() != "EOF" {
				return fmt.Errorf("error al leer la URL base: %w", err)
			}
			baseURL = strings.TrimSpace(urlInput)
			if baseURL == "" {
				baseURL = "https://openrouter.ai/api/v1/chat/completions"
			}

			fmt.Print("2. Modelo de IA [Por defecto: nvidia/nemotron-4-340b-instruct]: ")
			modelInput, err := reader.ReadString('\n')
			if err != nil && err.Error() != "EOF" {
				return fmt.Errorf("error al leer el modelo: %w", err)
			}
			model = strings.TrimSpace(modelInput)
			if model == "" {
				model = "nvidia/nemotron-4-340b-instruct"
			}

			fmt.Print("3. API Key (Token de autenticación obligatorio): ")
			keyInput, err := reader.ReadString('\n')
			if err != nil && err.Error() != "EOF" {
				return fmt.Errorf("error al leer el API Key: %w", err)
			}
			apiKey = strings.TrimSpace(keyInput)
			if apiKey == "" {
				return fmt.Errorf("el API Key es obligatorio para configurar Forge")
			}

			fmt.Print("4. Idioma para los commits (ej. 'es', 'en') [Por defecto: es]: ")
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
			fmt.Println("-------------------------------------------------------------------")
			fmt.Println("[OK] Configuración guardada exitosamente en ~/.forge.json")
			fmt.Println("===================================================================\n")
		}

		gitAdapter := git.NewGitAdapter()
		aiAdapter := ai.NewOpenAIAdapter(apiKey, baseURL, model)
		workflow := core.NewCommitWorkflow(gitAdapter, aiAdapter, cfg)

		// 2. Ejecución del workflow con retroalimentación visual al usuario
		fmt.Println("[INFO] Analizando cambios en staging...")
		proposal, err := workflow.Execute(cmd.Context())
		if err != nil {
			if errors.Is(err, core.ErrNoStagedChanges) {
				fmt.Println("[WARN] No hay cambios en staging. Utilice 'git add <archivos>' primero.")
				return nil
			}
			return fmt.Errorf("error al generar la propuesta de commit: %w", err)
		}

		// 3. Interactividad (Human-in-the-loop)
		fmt.Printf("\nPropuesta de Commit:\n\n%s\n\n", proposal.Subject)
		fmt.Print("¿Aceptar este commit? [Y/n]: ")

		input, err := reader.ReadString('\n')
		if err != nil && err.Error() != "EOF" {
			return fmt.Errorf("error al leer la confirmación del usuario: %w", err)
		}

		answer := strings.TrimSpace(input)
		if answer == "" || strings.EqualFold(answer, "y") {
			if err := gitAdapter.ExecuteCommit(cmd.Context(), proposal.Subject); err != nil {
				return fmt.Errorf("falló la confirmación en Git: %w", err)
			}
			fmt.Println("[OK] Commit creado exitosamente.")
		} else {
			fmt.Println("[INFO] Commit abortado por el usuario.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
}
