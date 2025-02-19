package util

import (
	"context"
	"log/slog"
	"time"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func Teardown(conf *config.Global) {
	if conf.Job != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		opts := metav1.DeleteOptions{PropagationPolicy: ptr.To(metav1.DeletePropagationForeground)}

		if viper.GetBool(consts.KeyCreateJob) {
			jobLog := slog.With("name", conf.Job.Name)
			jobLog.Info("Cleaning up job")
			if err := conf.Client.Jobs().Delete(ctx, conf.Job.Name, opts); err != nil {
				jobLog.Error("Failed to delete job", "error", err)
			}
		}

		if viper.GetBool(consts.KeyCreateNetworkPolicy) {
			netPolLog := slog.With("name", conf.Job.Name)
			netPolLog.Debug("Cleaning up network policy")
			if err := conf.Client.NetworkPolicies().Delete(ctx, conf.Job.Name, opts); err != nil {
				netPolLog.Error("Failed to delete network policy", "error", err)
			}
		}
	}
}
