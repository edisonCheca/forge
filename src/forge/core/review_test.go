package core_test

import (
	"testing"

	"github.com/forge/forge/src/forge/core"
)

func TestFilterDiff(t *testing.T) {
	rawDiff := `diff --git a/go.sum b/go.sum
index 1234..5678 100644
--- a/go.sum
+++ b/go.sum
+github.com/some/package v1.0.0
diff --git a/assets/logo.png b/assets/logo.png
Binary files differ
diff --git a/src/main.go b/src/main.go
index abcd..ef01 100644
--- a/src/main.go
+++ b/src/main.go
+func hello() string { return "world" }
`

	filtered, skipped := core.FilterDiff(rawDiff)

	if len(skipped) != 2 {
		t.Fatalf("expected 2 skipped files, got %d (%v)", len(skipped), skipped)
	}

	if !coreContains(filtered, "src/main.go") {
		t.Errorf("expected filtered diff to contain src/main.go, got: %s", filtered)
	}

	if coreContains(filtered, "go.sum") {
		t.Errorf("expected filtered diff to NOT contain go.sum")
	}

	if coreContains(filtered, "logo.png") {
		t.Errorf("expected filtered diff to NOT contain logo.png")
	}
}

func TestRunStaticAnalysis_Security(t *testing.T) {
	diffWithSecrets := `diff --git a/config.go b/config.go
--- a/config.go
+++ b/config.go
+const apiKey = "sk-1234567890abcdef1234567890"
+const ghToken = "ghp_1234567890abcdef1234567890"
diff --git a/.env b/.env
new file mode 100644
--- /dev/null
+++ b/.env
+DB_PASS=supersecret123
`

	findings := core.RunStaticAnalysis(diffWithSecrets)

	criticalCount := 0
	for _, f := range findings {
		if f.Severity == core.SeverityCritical {
			criticalCount++
		}
	}

	if criticalCount < 3 {
		t.Errorf("expected at least 3 critical findings for secrets and .env file, got %d", criticalCount)
	}
}

func TestRunStaticAnalysis_QualityWarnings(t *testing.T) {
	diffWithDebug := `diff --git a/src/app.js b/src/app.js
--- a/src/app.js
+++ b/src/app.js
+console.log("debugging user data:", user);
+debugger;
diff --git a/src/domain/calc.go b/src/domain/calc.go
--- a/src/domain/calc.go
+++ b/src/domain/calc.go
+fmt.Println("debug line in domain")
+panic("something went wrong")
`

	findings := core.RunStaticAnalysis(diffWithDebug)

	warningCount := 0
	for _, f := range findings {
		if f.Severity == core.SeverityWarning {
			warningCount++
		}
	}

	if warningCount < 4 {
		t.Errorf("expected at least 4 quality warnings (console.log, debugger, fmt.Println, panic), got %d", warningCount)
	}
}

func TestRunStaticAnalysis_Clean(t *testing.T) {
	cleanDiff := `diff --git a/src/service.go b/src/service.go
--- a/src/service.go
+++ b/src/service.go
+func ProcessOrder(order *Order) error {
+    if order == nil {
+        return errors.New("order cannot be nil")
+    }
+    return nil
+}
`

	findings := core.RunStaticAnalysis(cleanDiff)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for clean diff, got %d", len(findings))
	}
}

func coreContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && len(substr) > 0 && searchSubstr(s, substr))
}

func searchSubstr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
