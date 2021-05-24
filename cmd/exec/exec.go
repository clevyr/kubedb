package exec

import (
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/postgres"
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

var (
	database string
	username string
	password string
)

func init() {
	Command.Flags().StringVarP(&database, "dbname", "d", "db", "database name to connect to")
	Command.Flags().StringVarP(&username, "username", "U", "postgres", "database username")
	Command.Flags().StringVarP(&password, "password", "p", "", "database password")
}

func run(cmd *cobra.Command, args []string) (err error) {
	cmd.SilenceUsage = true

	client, err := kubernetes.CreateClientForCmd(cmd)
	if err != nil {
		return err
	}

	postgresPod, err := postgres.GetPod(client)
	if err != nil {
		return err
	}

	if password == "" {
		password, err = postgres.GetSecret(client)
		if err != nil {
			return err
		}
	}

	log.Println("Execing into \"" + postgresPod.Name + "\"")

	stdin := streams.NewIn(os.Stdin)
	if err := stdin.SetRawTerminal(); err != nil {
		return err
	}
	defer stdin.RestoreTerminal()

	return kubernetes.Exec(client, postgresPod, buildCommand(args), stdin, os.Stdout, stdin.IsTerminal())
}

func buildCommand(args []string) []string {
	var cmd []string
	if len(args) == 0 {
		cmd = []string{"PGPASSWORD=" + password, "psql", "--username=" + username, database}
	} else {
		cmd = append([]string{"PGPASSWORD=" + password, "exec"}, args...)
	}
	return []string{"sh", "-c", strings.Join(cmd, " ")}
}
