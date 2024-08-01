package cmd

import (
	"strings"

	"github.com/clevyr/kubedb/internal/database"
)

func newDescription() string {
	dbs := database.Names()

	return `Painlessly work with databases in Kubernetes.

Supported Databases:
  ` + strings.Join(dbs, ", ")
}
