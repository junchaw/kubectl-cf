package translations

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed trans.yaml
var transYAML []byte

//go:embed trans-fallback.yaml
var transFallbackYAML []byte

var trans map[string]string

// T is a wrapper for fmt.Sprintf, it will return the translated string
func T(messageID string, args ...any) string {
	return fmt.Sprintf(trans[messageID], args...)
}

func init() {
	if err := yaml.Unmarshal(transFallbackYAML, &trans); err != nil {
		panic(err)
	}
	var customized map[string]string
	if err := yaml.Unmarshal(transYAML, &customized); err != nil {
		panic(err)
	}
	for k, v := range customized {
		trans[k] = v
	}
}
