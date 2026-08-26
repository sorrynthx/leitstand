package highlight

import (
	"strings"
	"testing"
)

func TestHighlight(t *testing.T) {
	pyCode := `def hello():
    print("Hello, Leitstand!")
`
	highlighted := Highlight("test.py", pyCode)
	if highlighted == "" {
		t.Fatalf("Expected highlighted output, got empty")
	}

	yamlCode := `version: '3.8'
services:
  web:
    image: nginx:alpine
`
	highlightedYaml := Highlight("docker-compose.yaml", yamlCode)
	if !strings.Contains(highlightedYaml, "nginx") {
		t.Fatalf("Expected nginx in highlighted output")
	}
}
