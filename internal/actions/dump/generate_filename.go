package dump

import (
	"path/filepath"
	"time"
)

const DateFormat = "2006-01-02_150405"

type Filename struct {
	Database  string
	Namespace string
	Ext       string
	Date      time.Time
}

func (vars Filename) Generate() string {
	result := vars.Namespace + "_"
	// The database may be discovered from the cluster, so strip any directory
	// components to prevent the filename from escaping the output directory.
	switch database := filepath.Base(vars.Database); database {
	case "", ".", "..", string(filepath.Separator), vars.Namespace:
	default:
		result += database + "_"
	}
	result += vars.Date.Format(DateFormat) + vars.Ext
	return result
}
