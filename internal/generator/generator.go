// Package generator generates the files
package generator

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"text/template"

	"golang.org/x/exp/slog"

	"github.com/Djarvur/go-simple-converters/internal/extractor"
	"github.com/Djarvur/go-simple-converters/internal/parser"
)

const fileNameTmplSrc = `{{.Dir}}/simple_convert_codegen_{{.Name}}{{if .Test}}_test{{end}}.go`

//nolint:gochecknoglobals
var (
	//go:embed codegen.go.tmpl
	fileContentTmplSrc string

	fileContentTmpl = template.Must(template.New("fileContent").Parse(fileContentTmplSrc))
	fileNameTmpl    = template.Must(template.New("fileName").Parse(fileNameTmplSrc))
)

// Generate generates the files, calling parser and extractor.
func Generate(dir string, tags []string, includeTests bool, log *slog.Logger) error {
	pkgs, fset, err := parser.Parse(dir, tags, includeTests)
	if err != nil {
		return fmt.Errorf("parsing sources: %w", err)
	}

	for _, def := range extractor.Extract(pkgs, fset) {
		if err = writeFile(def); err != nil {
			return fmt.Errorf("generating: %w", err)
		}

		log.Debug("Generate", "def", def)
	}

	return nil
}

func buildFileName(data extractor.ConverterDef) (string, error) {
	var b bytes.Buffer

	if err := fileNameTmpl.Execute(&b, data); err != nil {
		return "", fmt.Errorf("%+v: %w", data, err)
	}

	return b.String(), nil
}

func writeFile(converterDef extractor.ConverterDef) error {
	fileName, err := buildFileName(converterDef)
	if err != nil {
		return fmt.Errorf("building file name: %w", err)
	}

	fileNameTmp := fileName + ".tmp"

	file, err := os.Create(fileName + ".tmp") //nolint:gosec
	if err != nil {
		return fmt.Errorf("opening file %q: %w", fileNameTmp, err)
	}

	defer file.Close() //nolint:errcheck

	if err = fileContentTmpl.Execute(file, converterDef); err != nil {
		return fmt.Errorf("writing file %q: %w", fileNameTmp, err)
	}

	if err = os.Rename(fileNameTmp, fileName); err != nil {
		return fmt.Errorf("renaming file %q to %q: %w", fileNameTmp, fileName, err)
	}

	return nil
}
