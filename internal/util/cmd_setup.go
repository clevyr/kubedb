package util

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/kubernetes"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
)

func DefaultFlags(cmd *cobra.Command, conf *config.Global) {
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

	grammarFlag, err := cmd.Flags().GetString("grammar")
	if err != nil {
		return err
	}

	var pods []v1.Pod
	if grammarFlag == "" {
		// Configure via detection
		conf.Grammar, pods, err = database.DetectGrammar(conf.Client)
		if err != nil {
			return err
		}
		log.WithField("grammar", conf.Grammar.Name()).Info("detected database grammar")
	} else {
		// Configure via flag
		conf.Grammar, err = database.New(grammarFlag)
		if err != nil {
			return err
		}
		log.WithField("grammar", conf.Grammar.Name()).Info("configured database grammar")

		pods, err = conf.Client.GetPodsFiltered(conf.Grammar.PodLabels())
		if err != nil {
			return err
		}
	}

	if len(pods) == 1 {
		conf.Pod = pods[0]
	} else {
		names := make([]string, 0, len(pods))
		for _, pod := range pods {
			names = append(names, pod.Name)
		}
		var idx int
		err = survey.AskOne(&survey.Select{
			Message: "Found multiple database pods. Select the desired instance.",
			Options: names,
		}, &idx)
		if err != nil {
			return err
		}
		conf.Pod = pods[idx]
	}

	if conf.Database == "" {
		conf.Database, err = conf.Client.GetValueFromEnv(conf.Pod, conf.Grammar.DatabaseEnvNames())
		if err != nil {
			conf.Database = conf.Grammar.DefaultDatabase()
			log.WithField("database", conf.Database).Warn("could not configure database from pod env, using default")
		} else {
			log.WithField("database", conf.Database).Info("configured database from pod env")
		}
	}

	if conf.Username == "" {
		conf.Username, err = conf.Client.GetValueFromEnv(conf.Pod, conf.Grammar.UserEnvNames())
		if err != nil {
			conf.Username = conf.Grammar.DefaultUser()
			log.WithField("user", conf.Username).Warn("could not configure user from pod env, using default")
		} else {
			log.WithField("user", conf.Username).Info("configured user from pod env")
		}
	}

	if conf.Password == "" {
		conf.Password, err = conf.Client.GetValueFromEnv(conf.Pod, conf.Grammar.PasswordEnvNames())
		if err != nil {
			return err
		}
	}

	return nil
}
