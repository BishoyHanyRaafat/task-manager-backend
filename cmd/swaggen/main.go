package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"task_manager/docs"
)

func main() {
	// 1️⃣ Get version from environment (set in CI)
	version := os.Getenv("SWAGGER_VERSION")
	if version == "" {
		version = "dev"
	}

	// 2️⃣ Update main.go (or docs.go) @version comment
	if err := updateSwaggerVersion("main.go", version); err != nil {
		fail(fmt.Errorf("updating swagger version failed: %w", err))
	}

	// Update the openapi.json
	createOpenapiJson()

	// 3️⃣ Run swag init
	args := []string{
		"init", "-g", "main.go", "--output", "docs",
		"--parseDependency", "--parseInternal",
	}
	swag, err := findSwag()
	if err != nil {
		fail(err)
	}

	cmd := exec.Command(swag, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fail(err)
	}
}

func createOpenapiJson() {
	_ = os.MkdirAll("docs", 0o755)
	_ = os.WriteFile("docs/openapi.json", []byte(docs.SwaggerInfo.ReadDoc()), 0o644)
}

// updateSwaggerVersion replaces the @version line in the file
func updateSwaggerVersion(filePath, version string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := bytes.Split(data, []byte("\n"))
	for i, line := range lines {
		if bytes.HasPrefix(bytes.TrimSpace(line), []byte("// @version")) {
			lines[i] = []byte("// @version " + version)
		}
	}

	return os.WriteFile(filePath, bytes.Join(lines, []byte("\n")), 0644)
}

func findSwag() (string, error) {
	// 1) PATH
	if p, err := exec.LookPath("swag"); err == nil {
		return p, nil
	}

	// 2) Look in GOBIN (modern Go)
	if gobin := os.Getenv("GOBIN"); gobin != "" {
		p := filepath.Join(gobin, "swag")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	// 3) GOPATH/bin/swag
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		// best-effort: ask `go env GOPATH`
		out, err := exec.Command("go", "env", "GOPATH").Output()
		if err == nil {
			gopath = string(bytesTrimSpace(out))
		}
	}

	if gopath != "" {
		p := filepath.Join(gopath, "bin", "swag")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("swag not found. Install it with: go install github.com/swaggo/swag/cmd/swag@latest")
}

func bytesTrimSpace(b []byte) []byte {
	i := 0
	j := len(b)
	for i < j && (b[i] == ' ' || b[i] == '\n' || b[i] == '\r' || b[i] == '\t') {
		i++
	}
	for j > i && (b[j-1] == ' ' || b[j-1] == '\n' || b[j-1] == '\r' || b[j-1] == '\t') {
		j--
	}
	return b[i:j]
}

func fail(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
