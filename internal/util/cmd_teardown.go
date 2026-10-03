package util

import (
	"context"
	"log/slog"
	"time"

	"github.com/clevyr/kubedb/internal/config/conftypes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Teardown deletes the job. Resources owned by the job, like the network policy and
// password secret, are garbage collected along with it.
func Teardown(conf *conftypes.Global) {
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
