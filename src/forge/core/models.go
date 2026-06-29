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
}

// CommitProposal representa la propuesta validada que el core devolverá al CLI.
type CommitProposal struct {
	Subject     string
	Body        string
	GeneratedAt time.Time
	ModelUsed   string
}
