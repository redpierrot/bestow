/*
All Rights Reversed (ɔ)
*/

package config

import (
	"bytes"
	"errors"
	"html/template"
	"testing"
)

const configFileTemplate = `# Bestow Configurations

# Config File Version
version: 0.1.0

# Bestow Profiles. Different profiles can have different source and destinations
profiles:
  default:
    # Source Directory (e.g. dotfiles repo)
    source: {{ .Src }}
    # Destination Directory (your $HOME directory)
    destination: {{ .Dest }}
`

func getConfigFileAsString(t *testing.T, src, dest string) string {
	t.Helper()
	tmpl, err := template.New("config").Parse(configFileTemplate)
	if err != nil {
		t.Fatal(err)
	}
	data := struct {
		Src  string
		Dest string
	}{
		Src:  src,
		Dest: dest,
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func validateErrScenario(t *testing.T, wantErr bool, err, wantErrIs error) bool {
	t.Helper()
	if (err != nil) != wantErr {
		t.Fatalf("got error %v, want %v", err, wantErr)
	}
	if wantErr && wantErrIs != nil && !errors.Is(err, wantErrIs) {
		t.Fatalf("error got %v, want %v", err, wantErrIs)
	}
	return wantErr
}
