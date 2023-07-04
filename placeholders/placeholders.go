package placeholders

import (
	"path/filepath"

	"github.com/barkimedes/go-deepcopy"
	api_v1 "github.com/vision-cli/api/v1"
	"github.com/vision-cli/common/transpiler/model"
)

const (
	ArgsCommandIndex = 0
	ArgsNameIndex    = 1
)

func SetupPlaceholders(req api_v1.PluginRequest) (*api_v1.PluginPlaceholders, error) {
	var err error
	p, err := deepcopy.Anything(&req.Placeholders)
	if err != nil {
		return nil, err
	}
	err = model.NameErr(req.Args[ArgsNameIndex])
	if err != nil {
		return nil, err
	}
	serviceName := req.Args[ArgsNameIndex]
	p.(*api_v1.PluginPlaceholders).ServiceName = serviceName
	p.(*api_v1.PluginPlaceholders).ServiceDirectory = filepath.Join(
		p.(*api_v1.PluginPlaceholders).ServicesDirectory,
		p.(*api_v1.PluginPlaceholders).ServiceNamespace,
		serviceName)
	return p.(*api_v1.PluginPlaceholders), nil
}
