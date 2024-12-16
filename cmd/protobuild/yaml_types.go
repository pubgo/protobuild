package protobuild

import (
	"github.com/pubgo/funk/assert"
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

type YamlListType[T any] []T

func (p *YamlListType[T]) UnmarshalYAML(value *yaml.Node) error {
	if value.IsZero() {
		return nil
	}

	switch value.Kind {
	case yaml.ScalarNode, yaml.MappingNode:
		var data T
		if err := value.Decode(&data); err != nil {
			return errors.WrapCaller(err)
		}
		*p = []T{data}
		return nil
	case yaml.SequenceNode:
		var data []T
		if err := value.Decode(&data); err != nil {
			return errors.WrapCaller(err)
		}
		*p = data
		return nil
	default:
		var val any
		assert.Exit(value.Decode(&val))
		return errors.Format("yaml kind type error, kind=%v data=%v", value.Kind, val)
	}
}

type strOrObject map[string]any

func (p *strOrObject) UnmarshalYAML(value *yaml.Node) error {
	if value.IsZero() {
		return nil
	}

	switch value.Kind {
	case yaml.ScalarNode:
		var data string
		if err := value.Decode(&data); err != nil {
			return errors.WrapCaller(err)
		}

		*p = strOrObject(map[string]any{"name": &data})
		return nil
	case yaml.MappingNode:
		var data map[string]any
		if err := value.Decode(&data); err != nil {
			return errors.WrapCaller(err)
		}

		*p = data
		return nil
	default:
		var val any
		assert.Exit(value.Decode(&val))
		return errors.Format("yaml kind type error,kind=%v data=%v", value.Kind, val)
	}
}
