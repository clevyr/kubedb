package exec

import (
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/docker/cli/cli/streams"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"log"
	"os"
	"strings"
)

var Command = &cobra.Command{
	Use:     "exec",
	Aliases: []string{"e", "shell"},
	Short:   "Connect to an interactive shell",
	RunE:    run,
}

var conf config.Exec

func init() {
	Command.Flags().StringVarP(&conf.Database, "dbname", "d", "", "database name to connect to")
	Command.Flags().StringVarP(&conf.Username, "username", "U", "", "database username")
	Command.Flags().StringVarP(&conf.Password, "password", "p", "", "database password")
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

	pod, err := client.GetPodByQueries(db.PodLabels())
	if err != nil {
		return err
	}

	if conf.Password == "" {
		conf.Password, err = client.GetSecretFromEnv(pod, db.PasswordEnvNames())
		if err != nil {
			return err
		}
	}

	log.Println("Exec into \"" + pod.Name + "\"")

	stdin := streams.NewIn(os.Stdin)
	if err := stdin.SetRawTerminal(); err != nil {
		return err
	}
	defer stdin.RestoreTerminal()

	return client.Exec(pod, buildCommand(db, conf, args), stdin, os.Stdout, stdin.IsTerminal())
}

func buildCommand(db database.Databaser, conf config.Exec, args []string) []string {
	var cmd []string
	if len(args) == 0 {
		cmd = db.ExecCommand(conf)
	} else {
		cmd = append([]string{"exec"}, args...)
	}
	return []string{"sh", "-c", strings.Join(cmd, " ")}
}
