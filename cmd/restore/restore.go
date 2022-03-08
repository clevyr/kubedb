package restore

import (
	"github.com/clevyr/kubedb/internal/cli"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/gzips"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/spf13/cobra"
	"io"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"log"
	"os"
	"strings"
)

var Command = &cobra.Command{
	Use:     "restore",
	Aliases: []string{"r"},
	Short:   "Restore a database",
	Args:    cobra.ExactArgs(1),
	PreRunE: preRun,
	RunE:    run,
}

var (
	conf        config.Restore
	inputFormat sqlformat.Format
)

func init() {
	Command.Flags().StringVarP(&conf.Database, "dbname", "d", "", "database name to connect to")
	Command.Flags().StringVarP(&conf.Username, "username", "U", "", "database username")
	Command.Flags().StringVarP(&conf.Password, "password", "p", "", "database password")

	Command.Flags().StringP("format", "F", "", "input format. inferred by default ([g]zip, [c]ustom, [p]lain text)")

	Command.Flags().BoolVarP(&conf.SingleTransaction, "single-transaction", "1", true, "restore as a single transaction")
	Command.Flags().BoolVarP(&conf.Clean, "clean", "c", true, "clean (drop) database objects before recreating")
	Command.Flags().BoolVarP(&conf.NoOwner, "no-owner", "O", true, "skip restoration of object ownership in plain-text format")

	Command.Flags().BoolVarP(&conf.Force, "force", "f", false, "do not prompt before restore")
}

func preRun(cmd *cobra.Command, args []string) error {
	formatStr, err := cmd.Flags().GetString("format")
	if err != nil {
		return err
	}
	inputFormat, err = sqlformat.ParseFormat(formatStr)
	if err != nil {
		inputFormat, err = sqlformat.ParseFilename(args[0])
		if err != nil {
			return err
		}
	}
	return nil
}

func run(cmd *cobra.Command, args []string) (err error) {
	cmd.SilenceUsage = true

	dbName, err := cmd.Flags().GetString("type")
	if err != nil {
		return err
	}
	db, err := database.New(dbName)

	if conf.Database == "" {
		conf.Database = db.DefaultDatabase()
	}
	if conf.Username == "" {
		conf.Username = db.DefaultUser()
	}

	client, err := kubernetes.CreateClientForCmd(cmd)
	if err != nil {
		return err
	}

	pod, err := db.GetPod(client)
	if err != nil {
		return err
	}

	if conf.Password == "" {
		conf.Password, err = db.GetSecret(client)
		if err != nil {
			return err
		}
	}

	f, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer f.Close()

	log.Println("Restoring \"" + args[0] + "\" to \"" + pod.Name + "\"")

	if !conf.Force {
		if err = cli.Confirm(os.Stdin, false); err != nil {
			return err
		}
	}

	var r io.Reader
	switch inputFormat {
	case sqlformat.Gzip, sqlformat.Custom:
		r = f
	case sqlformat.Plain:
		r = gzips.NewCompressReader(f)
	}

	if conf.Clean {
		resetReader := strings.NewReader(db.DropDatabaseQuery(conf.Database))
		err = kubernetes.Exec(client, pod, buildCommand(db, conf, sqlformat.Plain, false), resetReader, os.Stdout, false)
		if err != nil {
			return err
		}
	}

	err = kubernetes.Exec(client, pod, buildCommand(db, conf, inputFormat, true), r, os.Stdout, false)
	if err != nil {
		return err
	}

	err = kubernetes.Exec(client, pod, buildCommand(db, conf, sqlformat.Plain, false), strings.NewReader(db.AnalyzeQuery()), os.Stdout, false)
	if err != nil {
		return err
	}

	log.Println("Finished")
	return nil
}

func buildCommand(db database.Databaser, conf config.Restore, inputFormat sqlformat.Format, gunzip bool) []string {
	var cmd []string
	switch inputFormat {
	case sqlformat.Gzip, sqlformat.Plain:
		if gunzip {
			cmd = append([]string{"gunzip", "--force", "|"}, cmd...)
		}
	}
	cmd = append(cmd, db.RestoreCommand(conf, inputFormat)...)
	return []string{"sh", "-c", strings.Join(cmd, " ")}
}
