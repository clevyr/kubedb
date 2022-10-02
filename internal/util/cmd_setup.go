package util

import (
	"context"
	"errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database"
	"github.com/clevyr/kubedb/internal/kubernetes"
	"github.com/clevyr/kubedb/internal/log_hooks"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func DefaultSetup(cmd *cobra.Command, conf *config.Global) (err error) {
	cmd.SilenceUsage = true

	conf.Client, err = kubernetes.NewClient(conf.Kubeconfig, conf.Context, conf.Namespace)
	if err != nil {
		return err
	}
	log.WithField("namespace", conf.Client.Namespace).Debug("created kube client")
	conf.Namespace = conf.Client.Namespace

	access := config.NewNamespaceRegexp(cmd.Annotations["access"])
	if !access.Match(conf.Client.Namespace) {
		return errors.New("The current action is disabled for namespace " + conf.Client.Namespace)
	}

	dialectFlag, err := cmd.Flags().GetString("dialect")
	if err != nil {
		panic(err)
	}
	if dialectFlag == "" {
		dialectFlag, err = cmd.Flags().GetString("grammar")
		if err != nil {
			panic(err)
		}
	}

	var pods []v1.Pod
	if dialectFlag == "" {
		// Configure via detection
		conf.Dialect, pods, err = database.DetectDialect(conf.Client)
		if err != nil {
			return err
		}
		log.WithField("dialect", conf.Dialect.Name()).Debug("detected dialect")
	} else {
		// Configure via flag
		conf.Dialect, err = database.New(dialectFlag)
		if err != nil {
			return err
		}
		log.WithField("dialect", conf.Dialect.Name()).Debug("configured database")

		podFlag, err := cmd.Flags().GetString("pod")
		if err != nil {
			panic(err)
		}

		if podFlag == "" {
			pods, err = conf.Client.GetPodsFiltered(conf.Dialect.PodLabels())
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

	pods, err = conf.Dialect.FilterPods(conf.Client, pods)
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
		conf.Database, err = conf.Client.GetValueFromEnv(conf.Pod, conf.Dialect.DatabaseEnvNames())
		if err != nil {
			log.Debug("could not detect database from pod env")
		} else {
			log.WithField("database", conf.Database).Debug("found db name in pod env")
		}
	}

	conf.Username, err = cmd.Flags().GetString("username")
	if err != nil {
		panic(err)
	}
	if conf.Username == "" {
		conf.Username, err = conf.Client.GetValueFromEnv(conf.Pod, conf.Dialect.UserEnvNames())
		if err != nil {
			conf.Username = conf.Dialect.DefaultUser()
			log.WithField("user", conf.Username).Debug("could not detect user from pod env, using default")
		} else {
			log.WithField("user", conf.Username).Debug("found user in pod env")
		}
	}

	conf.Password, err = cmd.Flags().GetString("password")
	if err != nil {
		panic(err)
	}
	if conf.Password == "" {
		conf.Password, err = conf.Client.GetValueFromEnv(conf.Pod, conf.Dialect.PasswordEnvNames(*conf))
		if err != nil {
			return err
		}
	}
	if viper.GetBool("redact") {
		log.AddHook(log_hooks.Redact(conf.Password))
	}

	return nil
}
