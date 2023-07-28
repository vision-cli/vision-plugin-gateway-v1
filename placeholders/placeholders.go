package placeholders

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	api_v1 "github.com/vision-cli/api/v1"
)

const (
	ArgsCommandIndex = 0
	ArgsNameIndex    = 1
	// include any other arg indexes here
)

var nonAlphaRegex = regexp.MustCompile(`[^a-zA-Z]+`)

type Placeholders struct {
	RegistryServer string
	*api_v1.PluginPlaceholders
}

func SetupPlaceholders(req api_v1.PluginRequest) (*Placeholders, error) {
	var err error

	registryComponents := strings.Split(req.Placeholders.Registry, "/")
	if len(registryComponents) == 0 {
		return nil, fmt.Errorf("invalid registry server: %s", req.Placeholders.Registry)
	}
	p := &Placeholders{
		RegistryServer:     registryComponents[0],
		PluginPlaceholders: &req.Placeholders,
	}

	projectName := clearString(req.Args[ArgsNameIndex])
	p.ServiceName = projectName
	p.ServiceFqn, err = url.JoinPath(req.Placeholders.ServicesFqn, projectName)
	if err != nil {
		return nil, err
	}
	p.ProjectRoot = projectName
	p.ProjectName = projectName
	p.ProjectDirectory = projectName
	p.ProjectFqn, err = url.JoinPath(req.Placeholders.Remote, projectName)
	if err != nil {
		return nil, err
	}
	p.LibsFqn, err = url.JoinPath(req.Placeholders.Remote, projectName, "libs")
	if err != nil {
		return nil, err
	}
	return p, nil

}

func clearString(str string) string {
	return nonAlphaRegex.ReplaceAllString(str, "")
}
