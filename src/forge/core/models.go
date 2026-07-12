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
