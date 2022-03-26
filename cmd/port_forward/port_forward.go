package port_forward

import (
	"fmt"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/config/flags"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var Command = &cobra.Command{
	Use:     "port-forward [local_port]",
	Short:   "set up a local port forward",
	Args:    cobra.MaximumNArgs(1),
	RunE:    run,
	PreRunE: preRun,
}

var conf config.PortForward

func init() {
	flags.Address(Command, &conf.Addresses)
}

func preRun(cmd *cobra.Command, args []string) error {
	err := util.DefaultSetup(cmd, &conf.Global)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		conf.LocalPort = 30000 + conf.Grammar.DefaultPort()
	} else {
		port, err := strconv.ParseUint(args[0], 10, 16)
		if err != nil {
			return err
		}
		conf.LocalPort = uint16(port)
	}
	return nil
}

func run(cmd *cobra.Command, args []string) (err error) {
	log.WithField("pod", conf.Pod.Name).Info("setting up port forward")

	path := fmt.Sprintf(
		"/api/v1/namespaces/%s/pods/%s/portforward",
		conf.Pod.Namespace,
		conf.Pod.Name,
	)
	hostIP := strings.TrimLeft(conf.Client.ClientConfig.Host, "htps:/")

	transport, upgrader, err := spdy.RoundTripperFor(conf.Client.ClientConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})
	ports := []string{fmt.Sprintf("%d:%d", conf.LocalPort, conf.Grammar.DefaultPort())}
	stopCh := make(chan struct{}, 1)
	readyCh := make(chan struct{}, 1)
	fw, err := portforward.NewOnAddresses(dialer, conf.Addresses, ports, stopCh, readyCh, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}

	go func() {
		<-readyCh
		log.WithFields(log.Fields{
			"local":  conf.LocalPort,
			"remote": conf.Grammar.DefaultPort(),
		}).Info("port forward is ready")
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetTitle("%s Connection Parameters", conf.Pod.Namespace)
		t.AppendRows([]table.Row{
			{"Hostname", "localhost"},
			{"Port", conf.LocalPort},
			{"Username", conf.Username},
			{"Password", conf.Password},
			{"Database", conf.Database},
		})
		t.SetStyle(table.StyleLight)
		t.Render()
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
