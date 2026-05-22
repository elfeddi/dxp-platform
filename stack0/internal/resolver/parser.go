package resolver

import (
	"fmt"
	"os"

	"github.com/elfeddi/dxp/pkg/types"
	"gopkg.in/yaml.v3"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseFile(path string) (*types.DxPConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("lecture %s: %w", path, err)
	}
	return p.ParseBytes(data)
}

func (p *Parser) ParseBytes(data []byte) (*types.DxPConfig, error) {
	var config types.DxPConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}
	return &config, nil
}
