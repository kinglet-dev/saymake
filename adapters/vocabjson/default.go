package vocabjson

import (
	"bytes"
	_ "embed"

	"github.com/kinglet-dev/saymake/domain/vocab"
)

//go:embed default.json
var defaultJSON []byte

// Default returns the built-in vocabulary.
func Default() (vocab.Vocabulary, error) { return Load(bytes.NewReader(defaultJSON)) }
