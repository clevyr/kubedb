package util

import (
	"context"
	"time"

	"github.com/clevyr/kubedb/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Teardown(cmd *cobra.Command, conf *config.Global) {
	if conf.Job != nil {
		log.WithField("name", conf.Job.ObjectMeta.Name).Info("cleaning up job")
		policy := metav1.DeletePropagationForeground

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := conf.Client.Jobs().Delete(ctx, conf.Job.ObjectMeta.Name, metav1.DeleteOptions{
			PropagationPolicy: &policy,
		}); err != nil {
			log.WithField("name", conf.Job.ObjectMeta.Name).WithError(err).Error("failed to delete job")
		}
	}
}
