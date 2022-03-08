package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"io"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"log"
	"os"
)

var ErrNoPods = errors.New("no pods in namespace")

var ErrPodNotFound = errors.New("no pods with matching label")

func GetNamespacedPods(client KubeClient) (*v1.PodList, error) {
	pods, err := client.Pods().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return pods, err
	}

	if len(pods.Items) == 0 {
		return pods, fmt.Errorf("%w: %s", ErrNoPods, client.Namespace)
	}

	return pods, nil
}

func (client KubeClient) Exec(pod v1.Pod, command []string, stdin io.Reader, stdout io.Writer, tty bool) error {
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

func (client KubeClient) GetPodByQueries(queries []LabelQueryable) (v1.Pod, error) {
	pods, err := GetNamespacedPods(client)
	if err != nil {
		return v1.Pod{}, err
	}

	var errs []error
	for _, query := range queries {
		pod, err := query.FindPod(pods)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		return *pod, nil
	}

	if errors.Is(err, ErrPodNotFound) {
		for _, err := range errs {
			log.Println(err)
		}
		err = ErrPodNotFound
	}

	return v1.Pod{}, err
}
