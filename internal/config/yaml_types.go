// Package config provides YAML type helpers for configuration parsing.
package config

import (
	"github.com/pubgo/funk/errors"
	"gopkg.in/yaml.v3"
)

var _ yaml.Unmarshaler = (*PluginOpts)(nil)

// PluginOpts is a list of plugin options that can be unmarshaled from string or list.
type PluginOpts []string

// UnmarshalYAML implements yaml.Unmarshaler.
func (p *PluginOpts) UnmarshalYAML(value *yaml.Node) error {
	if value.IsZero() {
		return nil
	}

	switch value.Kind {
	case yaml.ScalarNode:
		if value.Value != "" {
			*p = []string{value.Value}
			return nil
		}
		return nil
	case yaml.SequenceNode:
		var data []string
		if err := value.Decode(&data); err != nil {
			return err
		}
		*p = data
		return nil
	default:
		return errors.Format("yaml kind type error, kind=%v data=%s", value.Kind, value.Value)
	}
}
