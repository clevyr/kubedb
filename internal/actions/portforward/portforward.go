package portforward

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/clevyr/kubedb/internal/config"
	kdblog "github.com/clevyr/kubedb/internal/log"
	"github.com/clevyr/kubedb/internal/tui"
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
		a.printTable()
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

//nolint:forbidigo
func (a PortForward) printTable() {
	log.Debug().
		Uint16("local", a.LocalPort).
		Uint16("remote", a.Port).
		Msg("port forward is ready")

	info := tui.MinimalTable(nil).
		RowIfNotEmpty("Context", a.Context).
		Row("Namespace", tui.NamespaceStyle(nil, a.Namespace).Render()).
		Row("Pod", a.DBPod.Name)

	params := tui.MinimalTable(nil).
		Row("Type", a.Dialect.PrettyName()).
		Row("Namespace", a.Namespace).
		Row("Hostname", "localhost").
		Row("Port", strconv.Itoa(int(a.LocalPort))).
		RowIfNotEmpty("Username", a.Username).
		RowIfNotEmpty("Password", a.Password).
		RowIfNotEmpty("Database", a.Database)

	tables := []*tui.Table{info, params}
	widths := make([]int, 0, len(tables))
	for _, t := range tables {
		widths = append(widths, lipgloss.Width(t.Render()))
	}
	widest := slices.Max(widths)
	differences := make([]int, 0, len(tables))
	for _, width := range widths {
		differences = append(differences, widest-width)
	}
	if slices.Max(differences) < 5 {
		for _, t := range tables {
			t.Width(widest)
		}
	}

	headerStyle := tui.HeaderStyle(nil)
	italicStyle := tui.TextStyle(nil).Italic(true)
	data := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinVertical(lipgloss.Center,
			headerStyle.Render("Database Instance"),
			info.Render(),
		),
		"",
		lipgloss.JoinVertical(lipgloss.Center,
			headerStyle.Render("Connection Parameters"),
			params.Render(),
		),
		"",
		headerStyle.Render("Tip:")+
			tui.TextStyle(nil).Render(" If you're connecting from a Docker container, set the hostname to ")+
			italicStyle.Render("host.docker.internal"),
	)

	baseStyle := lipgloss.NewStyle().
		Margin(1, 0).
		PaddingLeft(1).
		BorderStyle(lipgloss.ThickBorder()).
		BorderLeft(true).
		BorderForeground(lipgloss.Color("238"))
	fmt.Println(baseStyle.Render(data))
}
