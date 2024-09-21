package protobuild

import (
	"github.com/pubgo/funk/errors"
	yaml "gopkg.in/yaml.v3"
)

var _ yaml.Unmarshaler = (*pluginOpts)(nil)

type pluginOpts []string

func (p *pluginOpts) UnmarshalYAML(value *yaml.Node) error {
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
