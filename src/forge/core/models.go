package core

import "time"

// StagedFile representa un archivo modificado en el área de preparación (staging).
type StagedFile struct {
	Path      string
	Status    string // "M" (Modified), "A" (Added), "D" (Deleted), etc.
	Additions int
	Deletions int
}

// RepositoryContext encapsula la información extraída de Git necesaria para la IA.
type RepositoryContext struct {
	Branch        string
	StagedFiles   []StagedFile
	RawDiff       string
	RecentCommits []string
}

// GenerateRequest representa la solicitud formal enviada al motor de IA.
type GenerateRequest struct {
	Context            *RepositoryContext
	Language           string
	MaxLength          int
	ConventionalCommit bool
	IssueID            string
	ExtraContext       string
	Verbosity          string
}

// CommitProposal representa la propuesta validada que el core devolverá al CLI.
type CommitProposal struct {
	Subject     string
	Body        string
	GeneratedAt time.Time
	ModelUsed   string
}

// PRGenerateRequest encapsula el contexto de la rama y commits para generar un PR enriquecido por IA.
type PRGenerateRequest struct {
	Branch       string
	BaseBranch   string
	CommitLogs   []string
	Language     string
	StoryID      string
	StoryCode    string
	ExtraContext string
}

// PRProposal representa la propuesta de Pull Request generada por IA.
type PRProposal struct {
	Title       string
	Body        string
	GeneratedAt time.Time
	ModelUsed   string
}

// IssueLabel representa una etiqueta asociada a un issue en GitHub.
type IssueLabel struct {
	Name string `json:"name"`
}

// IssueInfo encapsula los detalles de una incidencia recuperada desde GitHub.
type IssueInfo struct {
	Number int          `json:"number"`
	Title  string       `json:"title"`
	Body   string       `json:"body"`
	Labels []IssueLabel `json:"labels"`
}

// ReviewSeverity indica el nivel de severidad de un hallazgo.
type ReviewSeverity string

const (
	SeverityCritical ReviewSeverity = "CRITICAL" // Secretos, archivos sensibles, tokens expuestos
	SeverityWarning  ReviewSeverity = "WARNING"  // Debuggers olvidados, prints, código de prueba
	SeverityInfo     ReviewSeverity = "INFO"     // Sugerencias de optimización o estilo
)

// ReviewFinding representa un hallazgo estático o semántico en el código.
type ReviewFinding struct {
	Severity ReviewSeverity
	Category string // "Security", "Quality", "Best Practices"
	File     string
	Line     int
	Message  string
	Snippet  string
}

// ReviewResult encapsula el reporte completo de la auditoría.
type ReviewResult struct {
	Passed       bool
	Findings     []ReviewFinding
	AISummary    string
	FilesChecked int
}

// ReviewRequest representa la solicitud de auditoría de código.
type ReviewRequest struct {
	Diff         string
	Language     string
	StaticOnly   bool
	ExtraContext string
}

