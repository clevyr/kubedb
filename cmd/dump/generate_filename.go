package dump

import (
	"fmt"
	"strings"
	"text/template"
	"time"
)

const DateFormat = "2006-01-02_150405"

var FilenameTemplate = fmt.Sprintf("{{ .Namespace }}_{{ .Date.Format %#v }}{{ .Ext }}", DateFormat)

type Filename struct {
	Namespace string
	Ext       string
	Date      time.Time
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
		Ext:       ".sql.gz",
		Date:      time.Date(2022, 1, 9, 9, 41, 0, 0, time.UTC),
	}.Generate()
	return filename
}
