package restore

import (
	"github.com/clevyr/kubedb/internal/command"
	"github.com/clevyr/kubedb/internal/config"
	"github.com/clevyr/kubedb/internal/database/sqlformat"
	"github.com/clevyr/kubedb/internal/progressbar"
	gzip "github.com/klauspost/pgzip"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
	"time"
)

type Restore struct {
	config.Restore `mapstructure:",squash"`
}

func (action Restore) Run() (err error) {
	var f io.ReadCloser
	switch action.Filename {
	case "-":
		f = os.Stdin
	default:
		f, err = os.Open(action.Filename)
		if err != nil {
			return err
		}
		defer func(f io.ReadCloser) {
			_ = f.Close()
		}(f)
	}

	pr, pw := io.Pipe()

	log.WithFields(log.Fields{
		"file":      action.Filename,
		"namespace": action.Client.Namespace,
		"pod":       action.Pod.Name,
	}).Info("ready to restore database")

	startTime := time.Now()

	bar := progressbar.New(-1, "uploading")
	errLog := progressbar.NewBarSafeLogger(os.Stderr, bar)
	outLog := progressbar.NewBarSafeLogger(os.Stdout, bar)
	log.SetOutput(errLog)

	ch := make(chan error, 1)
	go action.runInDatabasePod(pr, outLog, errLog, ch, action.Format)

	w := io.MultiWriter(pw, bar)

	if action.Clean && action.Format != sqlformat.Custom {
		dropQuery := action.Dialect.DropDatabaseQuery(action.Database)
		if dropQuery != "" {
			log.Info("cleaning existing data")
			err = gzipCopy(w, strings.NewReader(dropQuery))
			if err != nil {
				return err
			}
		}
	}

	log.Info("restoring database")
	switch action.Format {
	case sqlformat.Gzip, sqlformat.Custom, sqlformat.Unknown:
		_, err = io.Copy(w, f)
		if err != nil {
			return err
		}
	case sqlformat.Plain:
		err = gzipCopy(w, f)
		if err != nil {
			return err
		}
	}

	analyzeQuery := action.Dialect.AnalyzeQuery()
	if analyzeQuery != "" {
		if action.Format == sqlformat.Custom {
			_ = pw.Close()

			err = <-ch
			if err != nil {
				return err
			}

			pr, pw = io.Pipe()
			w = io.MultiWriter(pw, bar)
			go action.runInDatabasePod(pr, outLog, errLog, ch, sqlformat.Gzip)
		}

		log.Info("running analyze query")
		err = gzipCopy(w, strings.NewReader(analyzeQuery))
		if err != nil {
			return err
		}
	}

	_ = bar.Finish()
	log.SetOutput(os.Stderr)

	_ = pw.Close()

	err = <-ch
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"file": action.Filename,
		"in":   time.Since(startTime).Truncate(10 * time.Millisecond),
	}).Info("restore complete")
	return nil
}

func buildCommand(conf config.Restore, inputFormat sqlformat.Format) *command.Builder {
	cmd := conf.Dialect.RestoreCommand(conf, inputFormat).
		Unshift("gunzip", "--force", command.Pipe)
	log.WithField("cmd", cmd).Trace("finished building command")
	return cmd
}

func gzipCopy(w io.Writer, r io.Reader) (err error) {
	gzw := gzip.NewWriter(w)

	_, err = io.Copy(gzw, r)
	if err != nil {
		return err
	}

	err = gzw.Close()
	if err != nil {
		return err
	}

	return nil
}

func (action Restore) runInDatabasePod(r *io.PipeReader, stdout, stderr io.Writer, ch chan error, inputFormat sqlformat.Format) {
	var err error
	defer func(pr *io.PipeReader) {
		_ = pr.Close()
	}(r)

	err = action.Client.Exec(
		action.Pod,
		buildCommand(action.Restore, inputFormat).String(),
		r,
		stdout,
		stderr,
		false,
		nil,
	)
	if err != nil {
		_ = r.CloseWithError(err)
		ch <- err
		return
	}

	ch <- nil
}
