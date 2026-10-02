package util

import (
	"context"
	"log/slog"
	"time"

	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Teardown(conf *conftypes.Global) {
	deleteJob(conf)

	if conf.Job != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		foreground := metav1.DeletePropagationForeground
		opts := metav1.DeleteOptions{PropagationPolicy: &foreground}

		if config.Global.CreateNetworkPolicy {
			netPolLog := slog.With("name", conf.Job.Name)
			netPolLog.Debug("Cleaning up network policy")
			if err := conf.Client.NetworkPolicies().Delete(ctx, conf.Job.Name, opts); err != nil {
				netPolLog.Error("Failed to delete network policy", "error", err)
			}
		}
	}

	deleteJobSecret(conf)
}

func deleteJob(conf *conftypes.Global) {
	if conf.Job == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	foreground := metav1.DeletePropagationForeground
	opts := metav1.DeleteOptions{PropagationPolicy: &foreground}

	jobLog := slog.With("name", conf.Job.Name)
	jobLog.Info("Cleaning up job")
	if err := conf.Client.Jobs().Delete(ctx, conf.Job.Name, opts); err != nil {
		jobLog.Error("Failed to delete job", "error", err)
	}
}

func deleteJobSecret(conf *conftypes.Global) {
	if conf.JobSecret == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	secretLog := slog.With("name", conf.JobSecret.Name)
	secretLog.Debug("Cleaning up password secret")
	if err := conf.Client.Secrets().Delete(ctx, conf.JobSecret.Name, metav1.DeleteOptions{}); err != nil &&
		!apierrors.IsNotFound(err) {
		secretLog.Error("Failed to delete password secret", "error", err)
	}
}
