package gotests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// An expected error must skip only its own row, not the rest of the table.
func TestGenerateTestsExpectedErrorContinuesTable(t *testing.T) {
	for _, subtests := range []bool{false, true} {
		name := "without_subtests"
		if subtests {
			name = "with_subtests"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			src := filepath.Join(dir, "parse.go")
			code := "package example\nimport \"errors\"\nfunc Parse(n int) (int,error) { if n<0 { return 0,errors.New(\"negative\") }; return n,nil }\n"
			if err := os.WriteFile(src, []byte(code), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/regression\n\ngo 1.24.0\n"), 0600); err != nil {
				t.Fatal(err)
			}
			generated, err := GenerateTests(src, &Options{Subtests: subtests})
			if err != nil {
				t.Fatal(err)
			}
			output := strings.Replace(string(generated[0].Output), "// TODO: Add test cases.", `{name: "expected error", args: args{n: -1}, wantErr: true},
{name: "later result must be checked", args: args{n: 2}, want: 999},`, 1)
			if err := os.WriteFile(generated[0].Path, []byte(output), 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("go", "test", "-v", ".")
			cmd.Dir = dir
			result, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(result), "Parse() = 2, want 999") {
				t.Fatalf("later row was not checked: error=%v\n%s", err, result)
			}
		})
	}
}
