package dump

import (
	"bytes"
	"fmt"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"path/filepath"
	"text/template"
	"time"
)

const DateFormat = "2006-01-02_150405"

var FilenameTemplate = fmt.Sprintf("{{ .Namespace }}_{{ .Now %#v }}{{ .Ext }}", DateFormat)

type Filename struct {
	Dir       string
	Namespace string
	Format    sqlformat.Format
}

func (vars Filename) Now(layout string) string {
	return time.Now().Format(layout)
}

func (vars Filename) Ext() string {
	ext, err := sqlformat.WriteExtension(vars.Format)
	if err != nil {
		panic(err)
	}
	return ext
}

func (vars Filename) Generate() (string, error) {
	vars.Dir = filepath.Clean(vars.Dir)

	t, err := template.New("filename").Parse(FilenameTemplate)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, vars)
	if err != nil {
		return "", err
	}

	return filepath.Join(vars.Dir, tpl.String()), nil
}

func HelpFilename() string {
	filename, _ := Filename{
		Dir:       ".",
		Namespace: "clevyr",
		Format:    sqlformat.Gzip,
	}.Generate()
	return filename
}
