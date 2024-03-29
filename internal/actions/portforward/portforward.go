package portforward

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/clevyr/kubedb/internal/config"
	kdblog "github.com/clevyr/kubedb/internal/log"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PortForward struct {
	config.PortForward `mapstructure:",squash"`
}

func (a PortForward) Run(ctx context.Context) error {
	log.Debug().
		Str("namespace", a.Client.Namespace).
		Str("pod", a.DBPod.Name).
		Msg("setting up port forward")

	hostURL, err := url.Parse(a.Client.ClientConfig.Host)
	if err != nil {
		return err
	}
	hostURL.Path = path.Join("api", "v1", "namespaces", a.Client.Namespace, "pods", a.DBPod.Name, "portforward")

	transport, upgrader, err := spdy.RoundTripperFor(a.Client.ClientConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, hostURL)
	ports := []string{fmt.Sprintf("%d:%d", a.LocalPort, a.Port)}
	stopCh := make(chan struct{}, 1)
	readyCh := make(chan struct{}, 1)
	outWriter := kdblog.NewWriter(log.Logger, zerolog.InfoLevel)
	errWriter := kdblog.NewWriter(log.Logger, zerolog.ErrorLevel)
	fw, err := portforward.NewOnAddresses(dialer, a.Addresses, ports, stopCh, readyCh, outWriter, errWriter)
	if err != nil {
		return err
	}

	go func() {
		<-readyCh
		log.Debug().
			Uint16("local", a.LocalPort).
			Uint16("remote", a.Port).
			Msg("port forward is ready")
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetTitle(a.Namespace + " database")
		t.AppendRows([]table.Row{
			{"Type", a.Dialect.Name()},
			{"Hostname", "localhost"},
			{"Port", a.LocalPort},
		})
		if a.Username != "" {
			t.AppendRow(table.Row{"Username", a.Username})
		}
		if a.Password != "" {
			t.AppendRow(table.Row{"Password", a.Password})
		}
		if a.Database != "" {
			t.AppendRow(table.Row{"Database", a.Database})
		}
		t.SetStyle(table.StyleLight)
		t.Render()
		//nolint:forbidigo
		fmt.Println(`Tip: If you're connecting from a Docker container, set the hostname to "host.docker.internal"`)
	}()

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		<-ctx.Done()
		close(stopCh)
		return nil
	})

	group.Go(func() error {
		return fw.ForwardPorts()
	})

	return group.Wait()
}
