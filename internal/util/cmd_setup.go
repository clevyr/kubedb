package util

import (
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/kubernetes"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func DefaultFlags(cmd *cobra.Command, conf config.Global) {
	cmd.Flags().StringVarP(&conf.Database, "dbname", "d", "", "database name to connect to")
	cmd.Flags().StringVarP(&conf.Username, "username", "U", "", "database username")
	cmd.Flags().StringVarP(&conf.Password, "password", "p", "", "database password")

}

func DefaultSetup(cmd *cobra.Command, conf *config.Global) (err error) {
	cmd.SilenceUsage = true

	conf.Client, err = kubernetes.CreateClientForCmd(cmd)
	if err != nil {
		return err
	}
	log.WithField("namespace", conf.Client.Namespace).Info("created kube client")

	dbName, err := cmd.Flags().GetString("type")
	if err != nil {
		return err
	}
	conf.Databaser, err = database.New(dbName)
	if err != nil {
		return err
	}
	log.WithField("type", conf.Databaser.Name()).Info("configured database type")

	conf.Pod, err = conf.Client.GetPodByQueries(conf.Databaser.PodLabels())
	if err != nil {
		return err
	}

	if conf.Database == "" {
		conf.Database, err = conf.Client.GetValueFromEnv(conf.Pod, conf.Databaser.DatabaseEnvNames())
		if err != nil {
			conf.Database = conf.Databaser.DefaultDatabase()
			log.WithField("database", conf.Database).Warn("could not configure database from pod env, using default")
		} else {
			log.WithField("database", conf.Database).Info("configured database from pod env")
		}
	}

	if conf.Username == "" {
		conf.Username, err = conf.Client.GetValueFromEnv(conf.Pod, conf.Databaser.UserEnvNames())
		if err != nil {
			conf.Username = conf.Databaser.DefaultUser()
			log.WithField("user", conf.Username).Warn("could not configure user from pod env, using default")
		} else {
			log.WithField("user", conf.Username).Info("configured user from pod env")
		}
	}

	if conf.Password == "" {
		conf.Password, err = conf.Client.GetValueFromEnv(conf.Pod, conf.Databaser.PasswordEnvNames())
		if err != nil {
			return err
		}
	}

	return nil
}
