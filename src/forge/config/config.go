package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/forge/forge/src/forge/core"
)

// Verificación en tiempo de compilación para asegurar que AppConfig implementa core.ConfigPort.
var _ core.ConfigPort = (*AppConfig)(nil)

type fileConfig struct {
	BaseURL  string `json:"base_url"`
	Model    string `json:"model"`
	APIKey   string `json:"api_key"`
	Language string `json:"language"`
}

// AppConfig almacena la configuración de la aplicación leída desde archivo local o variables de entorno.
type AppConfig struct{}

// NewAppConfig crea una nueva instancia del gestor de configuración.
func NewAppConfig() *AppConfig {
	return &AppConfig{}
}

func (c *AppConfig) readConfigFile() (*fileConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(home, ".forge.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var fc fileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, err
	}
	return &fc, nil
}

// GetDefaultLanguage retorna el idioma configurado para los mensajes de commit (por defecto "en").
func (c *AppConfig) GetDefaultLanguage() string {
	if fc, err := c.readConfigFile(); err == nil && fc.Language != "" {
		return fc.Language
	}
	if lang := os.Getenv("FORGE_LANGUAGE"); lang != "" {
		return lang
	}
	return "en"
}

// RequireConventionalCommits determina si se debe exigir el formato Conventional Commits (por defecto true).
func (c *AppConfig) RequireConventionalCommits() bool {
	val := os.Getenv("FORGE_CONVENTIONAL")
	if val == "" {
		return true
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return true
	}
	return parsed
}

// GetMaxSubjectLength retorna la longitud máxima permitida para el subject del commit (72 por estándar de Git).
func (c *AppConfig) GetMaxSubjectLength() int {
	return 72
}

// GetAIBaseURL retorna la URL base de la API del proveedor de IA.
func (c *AppConfig) GetAIBaseURL() string {
	if fc, err := c.readConfigFile(); err == nil && fc.BaseURL != "" {
		return fc.BaseURL
	}
	if url := os.Getenv("FORGE_AI_BASE_URL"); url != "" {
		return url
	}
	return "https://api.openai.com/v1/chat/completions"
}

// GetSelectedAIModel retorna el modelo de IA seleccionado para generar propuestas.
func (c *AppConfig) GetSelectedAIModel() string {
	if fc, err := c.readConfigFile(); err == nil && fc.Model != "" {
		return fc.Model
	}
	if model := os.Getenv("FORGE_MODEL"); model != "" {
		return model
	}
	return "gpt-4o-mini"
}

// GetAIApiKey lee y retorna la clave de API necesaria para autenticarse con el proveedor de IA.
func (c *AppConfig) GetAIApiKey() string {
	if fc, err := c.readConfigFile(); err == nil && fc.APIKey != "" {
		return fc.APIKey
	}
	return os.Getenv("FORGE_AI_API_KEY")
}

// SaveConfig guarda de forma persistente la configuración en ~/.forge.json con permisos seguros.
func (c *AppConfig) SaveConfig(baseURL, model, apiKey, language string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error al obtener el directorio home del usuario: %w", err)
	}
	configPath := filepath.Join(home, ".forge.json")

	fc := fileConfig{
		BaseURL:  baseURL,
		Model:    model,
		APIKey:   apiKey,
		Language: language,
	}

	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return fmt.Errorf("error al serializar la configuración a JSON: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("error al escribir el archivo de configuración en %s: %w", configPath, err)
	}
	return nil
}
