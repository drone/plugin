package harness

import (
	"fmt"
	"html/template"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

// Metadata implements a source url generator that uses a template engine
// to generate url from give metadata.
type Metadata struct {
	tmpl    string
	funcMap template.FuncMap
}

// NewMetadata creates a new source Generator.
func NewMetadata(tmpl string, ref string) *Metadata {
	return &Metadata{
		tmpl: tmpl,
		funcMap: template.FuncMap{
			"arch":    func() string { return runtime.GOARCH },
			"os":      func() string { return runtime.GOOS },
			"release": releaseFunc(ref),
		},
	}
}

// Generate generates url from given template.
func (g *Metadata) Generate() (string, error) {
	sb := &strings.Builder{}

	t, err := template.New("source").Funcs(g.funcMap).Parse(g.tmpl)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("invalid source: %s", g.tmpl))
	}

	if err = t.Execute(sb, struct{}{}); err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("invalid source: %s", g.tmpl))
	}
	return sb.String(), nil
}

func releaseFunc(ref string) func() string {
	return func() string {
		if strings.HasPrefix(ref, "refs/tags/") {
			return strings.TrimPrefix(ref, "refs/tags/")
		}
		return "latest"
	}
}
