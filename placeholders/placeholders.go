package placeholders

import (
	"net/url"
	"regexp"

	"github.com/barkimedes/go-deepcopy"
	api_v1 "github.com/vision-cli/api/v1"
)

const (
	ArgsCommandIndex = 0
	ArgsNameIndex    = 1
	// include any other arg indexes here
)

var nonAlphaRegex = regexp.MustCompile(`[^a-zA-Z]+`)

type Placeholders struct {
	Name string
}

func SetupPlaceholders(req api_v1.PluginRequest) (*api_v1.PluginPlaceholders, error) {
	// setup your placeholders here
	// you can also deepcopy the Placeholders in the plugin request and use it
	// this is just an example:
	// name := clearString(req.Args[ArgsNameIndex])
	// return &Placeholders{
	// 	Name: name,
	// }, nil

	var err error
	p, err := deepcopy.Anything(&req.Placeholders)
	if err != nil {
		return nil, err
	}
	projectName := clearString(req.Args[ArgsNameIndex])
	p.(*api_v1.PluginPlaceholders).ProjectRoot = projectName
	p.(*api_v1.PluginPlaceholders).ProjectName = projectName
	p.(*api_v1.PluginPlaceholders).ProjectDirectory = projectName
	p.(*api_v1.PluginPlaceholders).ProjectFqn, err = url.JoinPath(req.Placeholders.Remote, projectName)
	if err != nil {
		return nil, err
	}
	p.(*api_v1.PluginPlaceholders).LibsFqn, err = url.JoinPath(req.Placeholders.Remote, projectName, "libs")
	if err != nil {
		return nil, err
	}
	return p.(*api_v1.PluginPlaceholders), nil
}

func clearString(str string) string {
	return nonAlphaRegex.ReplaceAllString(str, "")
}
