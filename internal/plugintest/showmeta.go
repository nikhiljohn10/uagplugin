package plugintest

import (
	"errors"
	"fmt"
	"plugin"

	"github.com/nikhiljohn10/uagplugin/models"
	"github.com/nikhiljohn10/uagplugin/typing"
)

func GetPluginMetadata(filePath string) (*models.MetaData, error) {
	p, err := plugin.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin file: %w", err)
	}

	sym, err := p.Lookup("Plugin")
	if err != nil {
		return nil, fmt.Errorf("failed to lookup 'Plugin' symbol: %w", err)
	}

	pl, ok := sym.(typing.Plugin)
	if !ok {
		if plPtr, ok := sym.(*typing.Plugin); ok && plPtr != nil {
			pl = *plPtr
		} else {
			return nil, errors.New("unexpected type for symbol 'Plugin', expected typing.Plugin")
		}
	}

	return pl.Meta(), nil
}
