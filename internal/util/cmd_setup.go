package util

import (
	"context"
	"github.com/AlecAivazis/survey/v2"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/kubernetes"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func DefaultSetup(cmd *cobra.Command, conf *config.Global) (err error) {
	cmd.SilenceUsage = true

	conf.Client, err = kubernetes.NewClientFromCmd(cmd)
	if err != nil {
		return err
	}
	log.WithField("namespace", conf.Client.Namespace).Info("created kube client")

	grammarFlag, err := cmd.Flags().GetString("grammar")
	if err != nil {
		panic(err)
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

		podFlag, err := cmd.Flags().GetString("pod")
		if err != nil {
			panic(err)
		}

		if podFlag == "" {
			pods, err = conf.Client.GetPodsFiltered(conf.Grammar.PodLabels())
			if err != nil {
				return err
			}
		} else {
			slashIdx := strings.IndexRune(podFlag, '/')
			if slashIdx != 0 && slashIdx+1 < len(podFlag) {
				podFlag = podFlag[slashIdx+1:]
			}
			pod, err := conf.Client.Pods().Get(context.Background(), podFlag, metav1.GetOptions{})
			if err != nil {
				return err
			}
			pods = []v1.Pod{*pod}
		}
	}

	pods, err = conf.Grammar.FilterPods(conf.Client, pods)
	if err != nil {
		log.WithError(err).Warn("could not query primary instance")
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

	conf.Database, err = cmd.Flags().GetString("dbname")
	if err != nil {
		panic(err)
	}
	if conf.Database == "" {
		conf.Database, err = conf.Client.GetValueFromEnv(conf.Pod, conf.Grammar.DatabaseEnvNames())
		if err != nil {
			conf.Database = conf.Grammar.DefaultDatabase()
			log.WithField("database", conf.Database).Warn("could not detect database from pod env, using default")
		} else {
			log.WithField("database", conf.Database).Info("found database from pod env")
		}
	}

	conf.Username, err = cmd.Flags().GetString("username")
	if err != nil {
		panic(err)
	}
	if conf.Username == "" {
		conf.Username, err = conf.Client.GetValueFromEnv(conf.Pod, conf.Grammar.UserEnvNames())
		if err != nil {
			conf.Username = conf.Grammar.DefaultUser()
			log.WithField("user", conf.Username).Warn("could not detect user from pod env, using default")
		} else {
			log.WithField("user", conf.Username).Info("found user in pod env")
		}
	}

	conf.Password, err = cmd.Flags().GetString("password")
	if err != nil {
		panic(err)
	}
	if conf.Password == "" {
		conf.Password, err = conf.Client.GetValueFromEnv(conf.Pod, conf.Grammar.PasswordEnvNames())
		if err != nil {
			return err
		}
	}

	return nil
}
