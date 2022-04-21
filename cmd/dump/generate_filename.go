package dump

import (
	"fmt"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"strings"
	"text/template"
	"time"
)

const DateFormat = "2006-01-02_150405"

var FilenameTemplate = fmt.Sprintf("{{ .Namespace }}_{{ .Date.Format %#v }}{{ .Ext }}", DateFormat)

type Filename struct {
	Namespace string
	Format    sqlformat.Format
	Date      time.Time
}

func (vars Filename) Ext() string {
	ext, err := sqlformat.WriteExtension(vars.Format)
	if err != nil {
		return ""
	}
	return ext
}

func (vars Filename) Generate() (string, error) {
	return generate(vars, FilenameTemplate)
}

func generate(vars Filename, tmpl string) (string, error) {
	t, err := template.New("filename").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	err = t.Execute(&buf, vars)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func HelpFilename() string {
	filename, _ := Filename{
		Namespace: "clevyr",
		Format:    sqlformat.Gzip,
		Date:      time.Date(2022, 1, 9, 9, 41, 0, 0, time.UTC),
	}.Generate()
	return filename
}
