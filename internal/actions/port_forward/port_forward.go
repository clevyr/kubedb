package port_forward

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/clevyr/kubedb/internal/config"
	log2 "github.com/clevyr/kubedb/internal/log"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PortForward struct {
	config.PortForward `mapstructure:",squash"`
}

func (a PortForward) Run(ctx context.Context) error {
	log.WithFields(log.Fields{
		"namespace": a.Client.Namespace,
		"pod":       a.Pod.Name,
	}).Info("setting up port forward")

	path := fmt.Sprintf(
		"/api/v1/namespaces/%s/pods/%s/portforward",
		a.Client.Namespace,
		a.Pod.Name,
	)
	hostIP := strings.TrimLeft(a.Client.ClientConfig.Host, "htps:/")

	transport, upgrader, err := spdy.RoundTripperFor(a.Client.ClientConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})
	ports := []string{fmt.Sprintf("%d:%d", a.LocalPort, a.Dialect.DefaultPort())}
	stopCh := make(chan struct{}, 1)
	readyCh := make(chan struct{}, 1)
	outWriter := log2.NewWriter(log.StandardLogger(), log.InfoLevel)
	errWriter := log2.NewWriter(log.StandardLogger(), log.ErrorLevel)
	fw, err := portforward.NewOnAddresses(dialer, a.Addresses, ports, stopCh, readyCh, outWriter, errWriter)
	if err != nil {
		return err
	}

	go func() {
		<-readyCh
		log.WithFields(log.Fields{
			"local":  a.LocalPort,
			"remote": a.Dialect.DefaultPort(),
		}).Info("port forward is ready")
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetTitle("%s Connection Parameters", a.Pod.Namespace)
		t.AppendRows([]table.Row{
			{"Type", a.Dialect.Name()},
			{"Hostname", "localhost"},
			{"Port", a.LocalPort},
			{"Username", a.Username},
			{"Password", a.Password},
		})
		if a.Database != "" {
			t.AppendRow(table.Row{"Database", a.Database})
		}
		t.SetStyle(table.StyleLight)
		t.Render()
		fmt.Println(`Tip: If you are connecting from a Docker container, try setting the hostname to "host.docker.internal"`)
	}()

	errCh := make(chan error, 1)
	go func() {
		errCh <- fw.ForwardPorts()
	}()

	interruptCh := make(chan os.Signal, 1)
	signal.Notify(interruptCh, os.Interrupt, syscall.SIGTERM)
	select {
	case err = <-errCh:
		if err != nil {
			return err
		}
	case <-interruptCh:
		log.Info("received exit signal")
		close(stopCh)
	}
	return nil
}
