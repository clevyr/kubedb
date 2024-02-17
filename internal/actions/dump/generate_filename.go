package dump

import (
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
	switch vars.Database {
	case "":
	case vars.Namespace:
	default:
		result += vars.Database + "_"
	}
	result += vars.Date.Format(DateFormat) + vars.Ext
	return result
}

func HelpFilename() string {
	return Filename{
		Namespace: "clevyr",
		Ext:       ".sql.gz",
		Date:      time.Date(2022, 1, 9, 9, 41, 0, 0, time.UTC),
	}.Generate()
}
