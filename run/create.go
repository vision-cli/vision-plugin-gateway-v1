package run

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	api_v1 "github.com/vision-cli/api/v1"

	"github.com/vision-cli/common/cases"
	"github.com/vision-cli/common/execute"
	"github.com/vision-cli/common/file"
	"github.com/vision-cli/common/module"
	"github.com/vision-cli/common/tmpl"
	"github.com/vision-cli/common/workspace"
	service "github.com/vision-cli/vision-plugin-service-v1/run"
	"github.com/vision-cli/vision-plugin-service-v1/svc"
)

const (
	goTemplateDir = "_templates/go"
	workflowDir   = ".github/workflows"
)

//go:embed all:_templates
var templateFiles embed.FS

func Create(p *api_v1.PluginPlaceholders, executor execute.Executor, t tmpl.TmplWriter) error {

	targetDir := filepath.Join(p.ServicesDirectory, p.ServiceNamespace, p.ServiceName)

	exposed, err := getExposedServices(p)
	if err != nil {
		return fmt.Errorf("finding exposed services: %w", err)
	}

	if err = createTemplate(p, targetDir, t); err != nil {
		return fmt.Errorf("generating service files with target dir: [%s]: %w", targetDir, err)
	}
	if err = generateGrpcHandlerCode(targetDir, exposed); err != nil {
		return fmt.Errorf("generating handler code with target dir: [%s]: %w", targetDir, err)
	}
	if err = generateModFiles(targetDir, p.ServiceFqn, exposed, executor, p); err != nil {
		return fmt.Errorf("generating module files with target dir: [%s]: %w", targetDir, err)
	}
	if err = GenWorkflow(p); err != nil {
		return fmt.Errorf("generating service workflow with target dir: [%s]: %w", targetDir, err)
	}
	if err = workspace.Use(".", targetDir, executor); err != nil {
		return fmt.Errorf("adding service to workspace: %w", err)
	}

	return nil
}

type serviceInfo struct {
	namespace   string
	serviceName string
	module      string
}

func getExposedServices(p *api_v1.PluginPlaceholders) ([]*serviceInfo, error) {
	exposed := make([]*serviceInfo, 0)

	err := svc.WalkAll(p.ServicesDirectory, func(fullPath, namespace, serviceName string) error {
		// skip the default namespace
		if namespace == "default" {
			return nil
		}
		protoPath := filepath.Join(fullPath, svc.ProtoDir)
		matches, err := fs.Glob(os.DirFS(protoPath), "*.gw.go")
		if err != nil {
			return fmt.Errorf("searching for .gw.go files in %s: %w", protoPath, err)
		}

		if len(matches) > 0 {
			module, err := module.Name(fullPath)
			if err != nil {
				return fmt.Errorf("finding service module in %s", fullPath)
			}

			exposed = append(exposed, &serviceInfo{
				namespace:   namespace,
				serviceName: serviceName,
				module:      module,
			})
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("searching for exposed services: %w", err)
	}

	return exposed, nil
}

func createTemplate(p *api_v1.PluginPlaceholders, targetDir string, t tmpl.TmplWriter) error {
	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("removing files from %s: %w", targetDir, err)
	}

	if err := tmpl.GenerateFS(templateFiles, goTemplateDir, targetDir, p, false, t); err != nil {
		return fmt.Errorf("generating the service structure from the template: %w", err)
	}

	return nil
}

func generateGrpcHandlerCode(targetDir string, exposed []*serviceInfo) error {
	if len(exposed) == 0 {
		return nil
	}

	handlersPath := filepath.Join(targetDir, svc.Handlers)
	lines, err := file.ToLines(handlersPath)
	if err != nil {
		return fmt.Errorf("reading handlers.go contents to slice of strings: %w", err)
	}

	configPath := filepath.Join(targetDir, "config", "config.go")
	configLines, err := file.ToLines(configPath)
	if err != nil {
		return fmt.Errorf("reading config.go contents to slice of strings: %w", err)
	}

	imports := []string{"\n"}
	body := []string{}
	configVars := []string{}
	for _, grpc := range exposed {
		alias := cases.Camel(grpc.namespace) + cases.Pascal(grpc.serviceName)
		pkgImport := filepath.Join(grpc.module, svc.ProtoDir)

		imports = append(imports, fmt.Sprintf(`	%sProto %q`, alias, pkgImport))

		configVars = append(configVars, fmt.Sprintf("	%s%sHost  string `envconfig:\"%s_%s_HOST\" default:\"0.0.0.0\"`",
			cases.Pascal(grpc.namespace), cases.Pascal(grpc.serviceName),
			strings.ToUpper(grpc.namespace), strings.ToUpper(grpc.serviceName),
		))

		body = append(body, fmt.Sprintf(
			`	if err := %sProto.Register%sHandlerFromEndpoint(ctx, mux, conf.%s%sHost+":"+conf.GrpcPort, opts); err != nil {
		return fmt.Errorf("failed to register gRPC service %s in namespace %s: %%w", err)
	}`,
			alias, cases.Pascal(grpc.serviceName), cases.Pascal(grpc.namespace), cases.Pascal(grpc.serviceName), grpc.serviceName, grpc.namespace))
	}
	body = append(body, "\n")

	lines = file.InsertIntoLines(lines, "google.golang.org/grpc", imports...)
	lines = file.InsertIntoLines(lines, "func Register", body...)
	configLines = file.InsertIntoLines(configLines, "type Config struct {", configVars...)

	if err = file.FromLines(handlersPath, lines); err != nil {
		return fmt.Errorf("writing handlers for exposed services: %w", err)
	}

	if err = file.FromLines(configPath, configLines); err != nil {
		return fmt.Errorf("writing config vars for exposed services: %w", err)
	}

	return nil
}

func generateModFiles(targetDir string, moduleName string, exposed []*serviceInfo, executor execute.Executor, p *api_v1.PluginPlaceholders) error {
	if err := module.Init(targetDir, moduleName, executor); err != nil {
		return fmt.Errorf("initialising module: %w", err)
	}
	if err := editModule(targetDir, exposed, executor, p); err != nil {
		return fmt.Errorf("editting module: %w", err)
	}
	if err := module.Tidy(targetDir, executor); err != nil {
		return fmt.Errorf("tidying module: %w", err)
	}
	return nil
}

func editModule(targetDir string, exposed []*serviceInfo, executor execute.Executor, p *api_v1.PluginPlaceholders) error {
	for _, grpc := range exposed {
		replacement := filepath.Join("../../..", p.ServicesDirectory, grpc.namespace, grpc.serviceName)

		if err := module.Replace(targetDir, grpc.module, replacement, executor); err != nil {
			return fmt.Errorf("editing go.mod file in %s: %w", targetDir, err)
		}
	}
	return nil
}

//go:embed _templates/workflows/go.yml.tmpl
var goWorkflow string

func GenWorkflow(p *api_v1.PluginPlaceholders) error {
	workflowName := svc.WorkflowName(p.ServiceNamespace, p.ServiceName)

	if err := service.Generate(goWorkflow, workflowDir, workflowName, p); err != nil {
		return fmt.Errorf("generating service workflow: %w", err)
	}

	return nil
}
