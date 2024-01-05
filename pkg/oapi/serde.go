package oapi

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type MapSlice[V any] []struct {
	Name  string
	Value *V
}

func (ms MapSlice[V]) Get(key string) (*V, bool) {
	for _, entry := range ms {
		if entry.Name == key {
			return entry.Value, true
		}
	}
	return nil, false
}

func (ms MapSlice[V]) With(key string, value *V) MapSlice[V] {
	for i, entry := range ms {
		if entry.Name == key {
			ms[i].Value = value
			return ms
		}
	}
	return append(ms, struct {
		Name  string
		Value *V
	}{
		Name:  key,
		Value: value,
	})
}

func (ms *MapSlice[V]) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("expected a mapping node, got %v", node.Kind)
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode, valueNode := node.Content[i], node.Content[i+1]
		var key string
		if err := keyNode.Decode(&key); err != nil {
			return err
		}

		var value V
		if err := valueNode.Decode(&value); err != nil {
			return err
		}

		*ms = append(*ms, struct {
			Name  string
			Value *V
		}{
			Name:  key,
			Value: &value,
		})
	}

	return nil
}
