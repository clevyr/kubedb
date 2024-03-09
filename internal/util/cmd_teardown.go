package util

import (
	"context"
	"time"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/consts"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func Teardown(conf *config.Global) {
	if conf.Job != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		opts := metav1.DeleteOptions{PropagationPolicy: ptr.To(metav1.DeletePropagationForeground)}

		log.WithField("name", conf.Job.ObjectMeta.Name).Info("cleaning up job")
		if err := conf.Client.Jobs().Delete(ctx, conf.Job.ObjectMeta.Name, opts); err != nil {
			log.WithField("name", conf.Job.ObjectMeta.Name).WithError(err).Error("failed to delete job")
		}

		if viper.GetBool(consts.CreateNetworkPolicyKey) {
			log.WithField("name", conf.Job.ObjectMeta.Name).Info("cleaning up network policy")
			if err := conf.Client.NetworkPolicies().Delete(ctx, conf.Job.ObjectMeta.Name, opts); err != nil {
				log.WithField("name", conf.Job.ObjectMeta.Name).WithError(err).Error("failed to delete network policy")
			}
		}
	}
}
