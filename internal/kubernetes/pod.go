package kubernetes

import (
	"context"
	"errors"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"os"
)

var PodNotFoundError = errors.New("could not find pod")

func GetPodByLabel(client KubeClient, key string, op selection.Operator, value []string) (pod v1.Pod, err error) {
	postgresReq, err := labels.NewRequirement(key, op, value)
	if err != nil {
		return v1.Pod{}, err
	}

	pods, err := client.Pods().List(context.TODO(), metav1.ListOptions{
		LabelSelector: labels.NewSelector().Add(*postgresReq).String(),
	})

	if len(pods.Items) > 0 {
		return pods.Items[0], err
	} else {
		return v1.Pod{}, PodNotFoundError
	}
}

func Exec(client KubeClient, pod v1.Pod, command []string, stdin io.Reader, stdout io.Writer, tty bool) error {
	req := client.ClientSet.CoreV1().RESTClient().Post().Resource("pods").Namespace(client.Namespace).
		Name(pod.Name).SubResource("exec")
	req.VersionedParams(&v1.PodExecOptions{
		Command: command,
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     tty,
	}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(client.ClientConfig, "POST", req.URL())
	if err != nil {
		return err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: os.Stderr,
	})

	return err
}
