// Package yamlpath contains an incomplete yamlpath implementation.
// TODO: Replace with github.com/caspr-io/yamlpath if it adds a license.
package yamlpath

import (
	"fmt"
	"strings"
)

// YAMLPath is a partial YAML Path implementation. It only supports slash
// separated lookup of nested keys.
func YAMLPath(data map[string]interface{}, key string) (interface{}, error) {
	p := strings.Split(strings.Trim(key, "/"), "/")
	if len(p) < 1 {
		return nil, fmt.Errorf("not found")
	}
	d, found := data[p[0]]
	if !found {
		return nil, fmt.Errorf("not found")
	}
	if len(p) < 2 {
		return d, nil
	}
	v, ok := d.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return YAMLPath(v, strings.Join(p[1:], "/"))
}
