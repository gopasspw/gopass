package secret

import (
	"bufio"
	"fmt"
	"sort"
	"strings"
)

func (s *Secret) decodeKV() error {
	mayBeYAML := false
	scanner := bufio.NewScanner(strings.NewReader(s.body))
	data := make(map[string]interface{}, strings.Count(s.body, "\n"))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "---") {
			mayBeYAML = true
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 1 {
			continue
		}
		if len(parts) == 1 && strings.HasPrefix(parts[0], "  ") {
			mayBeYAML = true
		}
		for i, part := range parts {
			parts[i] = strings.TrimSpace(part)
		}
		// preserve key only entries
		if len(parts) < 2 {
			data[parts[0]] = ""
			continue
		}
		if strings.HasPrefix(parts[1], "|") {
			mayBeYAML = true
		}
		data[parts[0]] = parts[1]
	}
	if mayBeYAML {
		docSep, err := s.decodeYAML()
		if debug {
			fmt.Printf("[DEBUG] decodeKV() - mayBeYAML - err: %s\n", err)
		}
		if docSep && err == nil && s.data != nil {
			return nil
		}
	}
	if debug {
		fmt.Printf("[DEBUG] decodeKV() - simple KV\n")
	}
	s.data = data
	return nil
}

func (s *Secret) encodeKV() error {
	if s.data == nil {
		return nil
	}
	keys := make([]string, 0, len(s.data))
	for key := range s.data {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var buf strings.Builder
	mayBeYAML := false
	for _, key := range keys {
		sv, ok := s.data[key].(string)
		if !ok {
			mayBeYAML = true
			continue
		}
		_, _ = buf.WriteString(key)
		_, _ = buf.WriteString(": ")
		_, _ = buf.WriteString(sv)
		_, _ = buf.WriteString("\n")
		if strings.Contains(sv, "\n") {
			mayBeYAML = true
		}
	}
	if mayBeYAML {
		if err := s.encodeYAML(); err == nil {
			if debug {
				fmt.Printf("[DEBUG] encodeKV() - mayBeYAML - OK\n")
			}
			return nil
		}
	}
	if debug {
		fmt.Printf("[DEBUG] encodeKV() - simple KV\n")
	}
	s.body = buf.String()
	return nil
}
