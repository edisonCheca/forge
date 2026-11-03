package core_test

import (
	"testing"

	"github.com/forge/forge/src/forge/core"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Control de Acceso (RBAC)", "control-de-acceso-rbac"},
		{"Autenticación y Sesión de Usuario", "autenticacion-y-sesion-de-usuario"},
		{"API REST: CRUD de Personal #10!", "api-rest-crud-de-personal-10"},
		{"   Espacios   extra   ", "espacios-extra"},
	}

	for _, tt := range tests {
		got := core.Slugify(tt.input)
		if got != tt.expected {
			t.Errorf("Slugify(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestBuildBranchName(t *testing.T) {
	tests := []struct {
		bType    string
		issueID  string
		huCode   string
		desc     string
		expected string
	}{
		{"feat", "5", "HU-01", "Control de Acceso RBAC", "feat/5-hu-01-control-de-acceso-rbac"},
		{"fix", "#12", "", "Error en login JWT", "fix/12-error-en-login-jwt"},
		{"docs", "3", "1", "Actualizar guías de usuario", "docs/3-hu-01-actualizar-guias-de-usuario"},
		{"", "7", "hu-2", "CRUD de Personal", "feat/7-hu-02-crud-de-personal"},
	}

	for _, tt := range tests {
		got := core.BuildBranchName(tt.bType, tt.issueID, tt.huCode, tt.desc)
		if got != tt.expected {
			t.Errorf("BuildBranchName(%q, %q, %q, %q) = %q; want %q", tt.bType, tt.issueID, tt.huCode, tt.desc, got, tt.expected)
		}
	}
}

func TestExtractBranchInfo(t *testing.T) {
	tests := []struct {
		branch       string
		wantStoryID  string
		wantStoryCode string
	}{
		{"feature/5-hu-01-rbac", "#5", "HU-01"},
		{"feat/3-hu-02-start-command", "#3", "HU-02"},
		{"fix/12-auth-bug", "#12", ""},
		{"docs/1-actualizar-arquitectura", "#1", ""},
		{"main", "", ""},
	}

	for _, tt := range tests {
		gotID, gotCode := core.ExtractBranchInfo(tt.branch)
		if gotID != tt.wantStoryID || gotCode != tt.wantStoryCode {
			t.Errorf("ExtractBranchInfo(%q) = (%q, %q); want (%q, %q)", tt.branch, gotID, gotCode, tt.wantStoryID, tt.wantStoryCode)
		}
	}
}

func TestParseIssueForBranch(t *testing.T) {
	issue1 := &core.IssueInfo{
		Number: 5,
		Title:  "[HU-01] Control de Acceso Basado en Roles (RBAC)",
		Labels: []core.IssueLabel{},
	}
	tType, tHU, tDesc := core.ParseIssueForBranch(issue1)
	if tType != "feat" || tHU != "HU-01" || tDesc != "Control de Acceso Basado en Roles (RBAC)" {
		t.Errorf("ParseIssueForBranch(issue1) = (%q, %q, %q); want (feat, HU-01, Control de Acceso Basado en Roles (RBAC))", tType, tHU, tDesc)
	}

	issue2 := &core.IssueInfo{
		Number: 12,
		Title:  "docs: actualizar guía de instalación",
		Labels: []core.IssueLabel{{Name: "documentation"}},
	}
	tType, tHU, tDesc = core.ParseIssueForBranch(issue2)
	if tType != "docs" || tHU != "" || tDesc != "actualizar guía de instalación" {
		t.Errorf("ParseIssueForBranch(issue2) = (%q, %q, %q); want (docs, \"\", actualizar guía de instalación)", tType, tHU, tDesc)
	}

	issue3 := &core.IssueInfo{
		Number: 8,
		Title:  "[FIX] Error de autenticación en JWT",
		Labels: []core.IssueLabel{{Name: "bug"}},
	}
	tType, tHU, tDesc = core.ParseIssueForBranch(issue3)
	if tType != "fix" || tDesc != "Error de autenticación en JWT" {
		t.Errorf("ParseIssueForBranch(issue3) = (%q, %q, %q); want (fix, \"\", Error de autenticación en JWT)", tType, tHU, tDesc)
	}
}

