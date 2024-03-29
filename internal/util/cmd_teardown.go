package util

import (
	"context"
	"time"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func Teardown(conf *config.Global) {
	if conf.Job != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		opts := metav1.DeleteOptions{PropagationPolicy: ptr.To(metav1.DeletePropagationForeground)}

		if viper.GetBool(consts.CreateJobKey) {
			jobLog := log.With().Str("name", conf.Job.Name).Logger()
			jobLog.Info().Msg("cleaning up job")
			if err := conf.Client.Jobs().Delete(ctx, conf.Job.Name, opts); err != nil {
				jobLog.Err(err).Msg("failed to delete job")
			}
		}

		if viper.GetBool(consts.CreateNetworkPolicyKey) {
			netPolLog := log.With().Str("name", conf.Job.Name).Logger()
			netPolLog.Info().Msg("cleaning up network policy")
			if err := conf.Client.NetworkPolicies().Delete(ctx, conf.Job.Name, opts); err != nil {
				netPolLog.Err(err).Msg("failed to delete network policy")
			}
		}
	}
}
