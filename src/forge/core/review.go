package core

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrNoDiffToReview es el error centinela cuando no existen cambios para auditar.
var ErrNoDiffToReview = errors.New("no changes found to review")

// Lista de archivos y extensiones a excluir en el diff para optimizar la ventana de contexto
var (
	excludedExtensions = []string{
		".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".webp", ".pdf",
		".woff", ".woff2", ".ttf", ".eot",
		".min.js", ".min.css", ".map",
		".exe", ".bin", ".tar", ".zip", ".gz",
	}

	excludedFiles = []string{
		"package-lock.json", "go.sum", "yarn.lock", "pnpm-lock.yaml",
		"composer.lock", "Cargo.lock", "Gemfile.lock",
	}

	sensitiveFilePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)^\.env(\..+)?$`),
		regexp.MustCompile(`(?i)\.(pem|key|pkcs12|pfx|p12)$`),
		regexp.MustCompile(`(?i)(id_rsa|id_ed25519|id_ecdsa)(\.pub)?$`),
		regexp.MustCompile(`(?i)(credentials|secrets|service_account)\.(json|yaml|yml)$`),
	}
)

// FilterDiff elimina archivos binarios, lockfiles y assets pesados de un diff para no saturar a la IA.
func FilterDiff(rawDiff string) (string, []string) {
	if rawDiff == "" {
		return "", nil
	}

	fileChunks := strings.Split(rawDiff, "diff --git ")
	var filteredChunks []string
	var skipped []string

	for _, chunk := range fileChunks {
		if strings.TrimSpace(chunk) == "" {
			continue
		}

		lines := strings.SplitN(chunk, "\n", 2)
		header := lines[0] // ej: "a/go.sum b/go.sum"

		shouldSkip := false
		for _, ef := range excludedFiles {
			if strings.Contains(header, ef) {
				shouldSkip = true
				skipped = append(skipped, ef)
				break
			}
		}

		if !shouldSkip {
			for _, ext := range excludedExtensions {
				if strings.Contains(strings.ToLower(header), ext) {
					shouldSkip = true
					skipped = append(skipped, ext)
					break
				}
			}
		}

		if !shouldSkip && strings.Contains(chunk, "Binary files ") {
			shouldSkip = true
			skipped = append(skipped, "binary file")
		}

		if !shouldSkip {
			// Limitar archivos excesivamente largos a 250 líneas
			allLines := strings.Split(chunk, "\n")
			if len(allLines) > 250 {
				truncated := strings.Join(allLines[:250], "\n")
				truncated += "\n... [diff truncado para proteger la ventana de contexto de la IA]\n"
				filteredChunks = append(filteredChunks, truncated)
			} else {
				filteredChunks = append(filteredChunks, chunk)
			}
		}
	}

	var result strings.Builder
	for _, fc := range filteredChunks {
		result.WriteString("diff --git ")
		result.WriteString(fc)
	}

	return result.String(), skipped
}

// RunStaticAnalysis audita un diff en busca de secretos expuestos y malas prácticas deterministas.
func RunStaticAnalysis(rawDiff string) []ReviewFinding {
	var findings []ReviewFinding
	if strings.TrimSpace(rawDiff) == "" {
		return findings
	}

	lines := strings.Split(rawDiff, "\n")
	currentFile := "desconocido"
	lineNum := 0

	// Patrones de secretos críticos
	reAPIKey := regexp.MustCompile(`sk-[a-zA-Z0-9_-]{20,}`)
	reGHPT := regexp.MustCompile(`(ghp|gho|ghu|ghs|github_pat)_[a-zA-Z0-9]{20,}`)
	reAWS := regexp.MustCompile(`AKIA[0-9A-Z]{16}`)
	rePrivKey := regexp.MustCompile(`-----BEGIN (RSA|EC|OPENSSH|PGP|PRIVATE) KEY-----`)
	reHardcodedSecret := regexp.MustCompile(`(?i)(password|passwd|api_key|secret_key|auth_token)\s*[:=]\s*["']([^"'\s]{8,})["']`)

	// Patrones de debug
	reJSDebug := regexp.MustCompile(`^\s*console\.(log|debug|warn|error)\(`)
	reJSBreak := regexp.MustCompile(`^\s*debugger\s*;?(\s*//.*)?$`)
	rePyPrint := regexp.MustCompile(`^\s*print\s*\(.*\)`)
	rePyBreak := regexp.MustCompile(`^\s*(breakpoint\(\)|pdb\.set_trace\(\))`)
	reGoPrint := regexp.MustCompile(`^\s*fmt\.(Println|Printf|Print)\(`)
	reGoPanic := regexp.MustCompile(`^\s*panic\(`)

	for _, line := range lines {
		// Detectar cambio de archivo
		if strings.HasPrefix(line, "diff --git ") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				currentFile = strings.TrimPrefix(parts[3], "b/")
			}
			lineNum = 0

			// 1. Verificar si el archivo en sí es sensible
			for _, pat := range sensitiveFilePatterns {
				partsOfPath := strings.Split(currentFile, "/")
				baseFileName := partsOfPath[len(partsOfPath)-1]
				if pat.MatchString(baseFileName) {
					findings = append(findings, ReviewFinding{
						Severity: SeverityCritical,
						Category: "Seguridad",
						File:     currentFile,
						Line:     1,
						Message:  fmt.Sprintf("Archivo sensible o de credenciales detectado en control de versiones: '%s'", baseFileName),
						Snippet:  currentFile,
					})
				}
			}
			continue
		}

		if strings.HasPrefix(line, "@@ ") {
			lineNum++
			continue
		}

		// Solo inspeccionamos líneas añadidas que no sean cabeceras de diff
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			lineNum++
			addedContent := strings.TrimPrefix(line, "+")
			trimmedContent := strings.TrimSpace(addedContent)

			// Ignorar comentarios
			if strings.HasPrefix(trimmedContent, "//") || strings.HasPrefix(trimmedContent, "#") || strings.HasPrefix(trimmedContent, "/*") {
				continue
			}

			isTestFile := strings.Contains(currentFile, "_test.") || strings.Contains(currentFile, ".test.") || strings.Contains(currentFile, ".spec.")
			isCLIFile := strings.Contains(currentFile, "/cli/") || strings.Contains(currentFile, "main.go")

			// Si es un archivo de pruebas, omitimos alertas de secretos y depuración
			if isTestFile {
				continue
			}

			// 2. Secretos y Tokens Críticos
			if rePrivKey.MatchString(trimmedContent) {
				findings = append(findings, ReviewFinding{
					Severity: SeverityCritical,
					Category: "Seguridad",
					File:     currentFile,
					Line:     lineNum,
					Message:  "Clave privada PEM/RSA/SSH expuesta en el código.",
					Snippet:  truncateSnippet(trimmedContent),
				})
			}

			if reAPIKey.MatchString(trimmedContent) {
				findings = append(findings, ReviewFinding{
					Severity: SeverityCritical,
					Category: "Seguridad",
					File:     currentFile,
					Line:     lineNum,
					Message:  "Posible API Key privada (formato sk-...) expuesta directamente en el código.",
					Snippet:  maskSecret(trimmedContent),
				})
			}

			if reGHPT.MatchString(trimmedContent) {
				findings = append(findings, ReviewFinding{
					Severity: SeverityCritical,
					Category: "Seguridad",
					File:     currentFile,
					Line:     lineNum,
					Message:  "Personal Access Token de GitHub expuesto.",
					Snippet:  maskSecret(trimmedContent),
				})
			}

			if reAWS.MatchString(trimmedContent) {
				findings = append(findings, ReviewFinding{
					Severity: SeverityCritical,
					Category: "Seguridad",
					File:     currentFile,
					Line:     lineNum,
					Message:  "Identificador de credencial de AWS (AKIA...) expuesto.",
					Snippet:  maskSecret(trimmedContent),
				})
			}

			if match := reHardcodedSecret.FindStringSubmatch(trimmedContent); len(match) > 2 {
				val := match[2]
				valLower := strings.ToLower(val)
				// Excluir placeholders inofensivos
				if !strings.Contains(valLower, "test") && !strings.Contains(valLower, "dummy") &&
					!strings.Contains(valLower, "password") && !strings.Contains(valLower, "your_") &&
					!strings.Contains(valLower, "changeme") && !strings.Contains(valLower, "default") {
					findings = append(findings, ReviewFinding{
						Severity: SeverityCritical,
						Category: "Seguridad",
						File:     currentFile,
						Line:     lineNum,
						Message:  fmt.Sprintf("Posible secreto o contraseña hardcodeada ('%s'). Use variables de entorno.", match[1]),
						Snippet:  maskSecret(trimmedContent),
					})
				}
			}

			// 3. Sentencias de Debug Olvidadas
			if reJSDebug.MatchString(trimmedContent) {
				findings = append(findings, ReviewFinding{
					Severity: SeverityWarning,
					Category: "Calidad",
					File:     currentFile,
					Line:     lineNum,
					Message:  "Llamada 'console.log/debug' detectada en código productivo.",
					Snippet:  truncateSnippet(trimmedContent),
				})
			}

			if reJSBreak.MatchString(trimmedContent) {
				findings = append(findings, ReviewFinding{
					Severity: SeverityWarning,
					Category: "Calidad",
					File:     currentFile,
					Line:     lineNum,
					Message:  "Punto de interrupción 'debugger' olvidado.",
					Snippet:  truncateSnippet(trimmedContent),
				})
			}

			if reGoPrint.MatchString(trimmedContent) && !isCLIFile {
				findings = append(findings, ReviewFinding{
					Severity: SeverityWarning,
					Category: "Calidad",
					File:     currentFile,
					Line:     lineNum,
					Message:  "Llamada a 'fmt.Print*' en paquete de lógica/dominio. Utilice un logger estructurado.",
					Snippet:  truncateSnippet(trimmedContent),
				})
			}

			if reGoPanic.MatchString(trimmedContent) {
				findings = append(findings, ReviewFinding{
					Severity: SeverityWarning,
					Category: "Calidad",
					File:     currentFile,
					Line:     lineNum,
					Message:  "Uso de 'panic()' en código productivo. Prefiera retornar errores controlados.",
					Snippet:  truncateSnippet(trimmedContent),
				})
			}

			if rePyPrint.MatchString(trimmedContent) && !isCLIFile {
				findings = append(findings, ReviewFinding{
					Severity: SeverityWarning,
					Category: "Calidad",
					File:     currentFile,
					Line:     lineNum,
					Message:  "Sentencia 'print()' detectada en módulo Python. Prefiera el módulo 'logging'.",
					Snippet:  truncateSnippet(trimmedContent),
				})
			}

			if rePyBreak.MatchString(trimmedContent) {
				findings = append(findings, ReviewFinding{
					Severity: SeverityWarning,
					Category: "Calidad",
					File:     currentFile,
					Line:     lineNum,
					Message:  "Punto de interrupción 'breakpoint()' o 'pdb' detectado en Python.",
					Snippet:  truncateSnippet(trimmedContent),
				})
			}
		}
	}

	return findings
}

func truncateSnippet(s string) string {
	if len(s) > 80 {
		return s[:77] + "..."
	}
	return s
}

func maskSecret(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}

// ReviewWorkflow orquesta la auditoría estática y semántica con IA.
type ReviewWorkflow struct {
	git    GitPort
	ai     AIPort
	config ConfigPort
}

// NewReviewWorkflow crea una nueva instancia de ReviewWorkflow.
func NewReviewWorkflow(git GitPort, ai AIPort, config ConfigPort) *ReviewWorkflow {
	return &ReviewWorkflow{
		git:    git,
		ai:     ai,
		config: config,
	}
}

// Execute ejecuta la revisión de código sobre staging o la rama activa.
func (w *ReviewWorkflow) Execute(ctx context.Context, branchReview bool, baseBranch string, staticOnly bool, extraContext string) (*ReviewResult, error) {
	var rawDiff string
	var err error

	if branchReview {
		rawDiff, err = w.git.GetBranchDiff(ctx, baseBranch)
	} else {
		rawDiff, err = w.git.GetStagedDiff(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("error al obtener diff para auditoría: %w", err)
	}

	if strings.TrimSpace(rawDiff) == "" {
		return nil, ErrNoDiffToReview
	}

	// 1. Análisis Estático Determinista (0 tokens)
	findings := RunStaticAnalysis(rawDiff)

	// 2. Pre-filtrado de Diff para proteger la ventana de contexto
	filteredDiff, _ := FilterDiff(rawDiff)

	// Contar archivos revisados
	filesChecked := strings.Count(rawDiff, "diff --git ")

	result := &ReviewResult{
		Findings:     findings,
		FilesChecked: filesChecked,
		Passed:       true,
	}

	// Si hay hallazgos críticos de seguridad, la revisión se considera no aprobada
	for _, f := range findings {
		if f.Severity == SeverityCritical {
			result.Passed = false
			break
		}
	}

	// 3. Análisis Semántico con IA (Opcional, si no es staticOnly y existe API key)
	if !staticOnly && w.ai != nil && w.config.GetAIApiKey() != "" && strings.TrimSpace(filteredDiff) != "" {
		req := &ReviewRequest{
			Diff:         filteredDiff,
			Language:     w.config.GetDefaultLanguage(),
			StaticOnly:   false,
			ExtraContext: extraContext,
		}

		aiSummary, errAI := w.ai.ReviewCode(ctx, req)
		if errAI == nil && strings.TrimSpace(aiSummary) != "" {
			result.AISummary = aiSummary
		}
	}

	return result, nil
}
