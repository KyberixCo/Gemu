package steam

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type node struct {
	value    string
	children map[string]node
}

// tokenize handles quoted strings, escaped paths, braces and Valve comments.
func tokenize(s string) ([]string, error) {
	var tokens []string
	for i := 0; i < len(s); {
		if unicode.IsSpace(rune(s[i])) {
			i++
			continue
		}
		if strings.HasPrefix(s[i:], "//") {
			for i < len(s) && s[i] != '\n' {
				i++
			}
			continue
		}
		if s[i] == '{' || s[i] == '}' {
			tokens = append(tokens, string(s[i]))
			i++
			continue
		}
		if s[i] == '"' {
			start := i
			i++
			for i < len(s) && s[i] != '"' {
				if s[i] == '\\' {
					i++
				}
				i++
			}
			if i >= len(s) {
				return nil, fmt.Errorf("unterminated VDF string")
			}
			value, err := strconv.Unquote(s[start : i+1])
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, value)
			i++
			continue
		}
		start := i
		for i < len(s) && !unicode.IsSpace(rune(s[i])) && s[i] != '{' && s[i] != '}' {
			i++
		}
		tokens = append(tokens, s[start:i])
	}
	return tokens, nil
}
func parse(s string) (map[string]node, error) {
	tokens, err := tokenize(strings.TrimPrefix(s, "\ufeff"))
	if err != nil {
		return nil, err
	}
	i := 0
	var block func(bool) (map[string]node, error)
	block = func(nested bool) (map[string]node, error) {
		out := map[string]node{}
		for i < len(tokens) {
			key := tokens[i]
			i++
			if key == "}" {
				if !nested {
					return nil, fmt.Errorf("unexpected closing brace")
				}
				return out, nil
			}
			if key == "{" || i >= len(tokens) {
				return nil, fmt.Errorf("missing value for %q", key)
			}
			value := tokens[i]
			i++
			if value == "{" {
				children, err := block(true)
				if err != nil {
					return nil, err
				}
				out[strings.ToLower(key)] = node{children: children}
			} else if value == "}" {
				return nil, fmt.Errorf("missing value for %q", key)
			} else {
				out[strings.ToLower(key)] = node{value: value}
			}
		}
		if nested {
			return nil, fmt.Errorf("unclosed VDF block")
		}
		return out, nil
	}
	return block(false)
}
