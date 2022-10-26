package kubernetes

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

var ErrNoPods = errors.New("no pods in namespace")

var ErrPodNotFound = errors.New("no pods with matching label")

func (client KubeClient) GetNamespacedPods(ctx context.Context) (*v1.PodList, error) {
	pods, err := client.Pods().List(ctx, metav1.ListOptions{})
	if err != nil {
		return pods, err
	}

	if len(pods.Items) == 0 {
		return pods, fmt.Errorf("%w: %s", ErrNoPods, client.Namespace)
	}

	return pods, nil
}

func (client KubeClient) Exec(pod v1.Pod, cmd string, stdin io.Reader, stdout, stderr io.Writer, tty bool, terminalSizeQueue remotecommand.TerminalSizeQueue) error {
	req := client.ClientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(client.Namespace).
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Command: []string{"sh", "-c", cmd},
			Stdin:   stdin != nil,
			Stdout:  stdout != nil,
			Stderr:  stderr != nil,
			TTY:     tty,
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(client.ClientConfig, "POST", req.URL())
	if err != nil {
		return err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            stdout,
		Stderr:            stderr,
		Tty:               tty,
		TerminalSizeQueue: terminalSizeQueue,
	})

	return err
}

func (client KubeClient) GetPodsFiltered(ctx context.Context, queries []LabelQueryable) ([]v1.Pod, error) {
	pods, err := client.GetNamespacedPods(ctx)
	if err != nil {
		return []v1.Pod{}, err
	}
	return client.FilterPodList(pods, queries)
}

func (client KubeClient) FilterPodList(pods *v1.PodList, queries []LabelQueryable) (foundPods []v1.Pod, err error) {
	for _, query := range queries {
		var p []v1.Pod
		p, err = query.FindPods(pods)
		if errors.Is(err, ErrPodNotFound) {
			log.WithField("query", query).Trace(err)
			continue
		}
		log.WithFields(log.Fields{
			"query": query,
			"count": len(p),
		}).Trace("query returned podlist")
		foundPods = append(foundPods, p...)
	}

	if len(foundPods) == 0 {
		if errors.Is(err, ErrPodNotFound) {
			err = ErrPodNotFound
		}

		return []v1.Pod{}, err
	}
	return foundPods, nil
}
