package run

import (
	"embed"
	"fmt"

	"github.com/vision-cli/common/execute"
	"github.com/vision-cli/common/tmpl"
	"github.com/vision-cli/vision-plugin-gateway-v1/placeholders"
)

const (
	goTemplateDir = "_templates"
)

//go:embed all:_templates
var templateFiles embed.FS

func Create(p *placeholders.Placeholders, executor execute.Executor, t tmpl.TmplWriter) error {
	var err error

	if err = tmpl.GenerateFS(templateFiles, goTemplateDir, p.Name, p, false, t); err != nil {
		return fmt.Errorf("generating structure from the template: %w", err)
	}

	return nil
}
