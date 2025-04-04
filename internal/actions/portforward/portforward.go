package portforward

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"slices"
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/database/postgres"
	"github.com/clevyr/kubedb/internal/log"
	"github.com/clevyr/kubedb/internal/tui"
	"golang.org/x/time/rate"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PortForward struct {
	conftypes.PortForward `koanf:",squash"`
}

func (a PortForward) Run(ctx context.Context) error {
	slog.Debug("Setting up port forward",
		"namespace", a.Client.Namespace,
		"pod", a.DBPod.Name,
	)

	hostURL, err := url.Parse(a.Client.ClientConfig.Host)
	if err != nil {
		return err
	}
	hostURL.Path = path.Join("api", "v1", "namespaces", a.Client.Namespace, "pods", a.DBPod.Name, "portforward")

	transport, upgrader, err := spdy.RoundTripperFor(a.Client.ClientConfig)
	if err != nil {
		return err
	}

	readyCh, stopCh := make(chan struct{}), make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
		case <-readyCh:
			readyCh = nil
			a.printTable()
		}
	}()

	go func() {
		<-ctx.Done()
		close(stopCh)
	}()

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, hostURL)
	ports := []string{fmt.Sprintf("%d:%d", a.ListenPort, a.Port)}
	outWriter := log.NewWriter(slog.Default(), slog.LevelInfo)
	errWriter := log.NewWriter(slog.Default(), slog.LevelError)

	limiter := rate.NewLimiter(rate.Every(time.Second), 3)
	for {
		fw, err := portforward.NewOnAddresses(dialer, a.Address, ports, stopCh, readyCh, outWriter, errWriter)
		if err != nil {
			return err
		}

		err = fw.ForwardPorts()
		switch {
		case err == nil:
			return nil
		case ctx.Err() != nil:
			return ctx.Err()
		case errors.Is(err, portforward.ErrLostConnectionToPod):
		default:
			return err
		}

		if err := limiter.Wait(ctx); err != nil {
			return err
		}
		slog.Info("Reconnecting")
	}
}

func (a PortForward) printTable() {
	slog.Debug("Port forward is ready",
		"local", a.ListenPort,
		"remote", a.Port,
	)

	info := tui.MinimalTable(nil).
		RowIfNotEmpty("Context", a.Context).
		Row("Namespace", tui.NamespaceStyle(nil, a.NamespaceColors, a.Namespace).Render()).
		Row("Pod", a.DBPod.Name)

	params := tui.MinimalTable(nil).
		Row("Type", a.Dialect.PrettyName()).
		Row("Namespace", a.Namespace).
		Row("Hostname", "localhost").
		Row("Port", strconv.Itoa(int(a.ListenPort))).
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

	tips := list.New(
		tui.TextStyle(nil).Render("To connect from a Docker container, set the hostname to ") +
			italicStyle.Render("host.docker.internal"),
	).Enumerator(func(list.Items, int) string {
		return " •"
	})

	if _, ok := a.Dialect.(postgres.Postgres); ok {
		tips.Item(
			tui.TextStyle(nil).Render("Postgres causes reconnects when SSL is enabled. Disable SSL by adding ") +
				italicStyle.Render("sslmode=disable") + " to your connection string",
		)
	}

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
		headerStyle.Render("Tips:"),
		tips.String(),
	)

	baseStyle := lipgloss.NewStyle().
		Margin(1, 0).
		PaddingLeft(1).
		BorderStyle(lipgloss.ThickBorder()).
		BorderLeft(true).
		BorderForeground(lipgloss.Color("238"))
	_, _ = fmt.Fprintln(os.Stdout, baseStyle.Render(data))
}
