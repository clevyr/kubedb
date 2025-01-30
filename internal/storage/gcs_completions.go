package storage

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"gabe565.com/utils/bytefmt"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
)

func CompleteBucketsGCS(u *url.URL, projectID string) ([]string, cobra.ShellCompDirective) {
	if projectID == "" {
		if val := os.Getenv("GOOGLE_CLOUD_PROJECT"); val != "" {
			projectID = val
		} else if val := os.Getenv("GCLOUD_PROJECT"); val != "" {
			projectID = val
		} else if val := os.Getenv("GCP_PROJECT"); val != "" {
			projectID = val
		}
	}

	u.Path = "/"

	buckets, count, err := ListBucketsGCS(context.Background(), projectID)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	names := make([]string, 0, count)
	for bucket, err := range buckets {
		if err != nil {
			slog.Error("Failed to list GCS buckets", "error", err)
			return nil, cobra.ShellCompDirectiveError
		}

		u.Host = bucket.Name
		names = append(names, u.String())
	}
	return names, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
}

func CompleteObjectsGCS(u *url.URL, exts []string, dirOnly bool) ([]string, cobra.ShellCompDirective) {
	objects, count, err := ListObjectsGCS(context.Background(), u.String())
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	paths := make([]string, 0, count)
	for object, err := range objects {
		if err != nil {
			slog.Error("Failed to list GCS objects", "error", err)
			return nil, cobra.ShellCompDirectiveError
		}

		if object.Prefix != "" {
			u.Path = object.Prefix
			paths = append(paths, u.String())
		} else if !dirOnly && util.FilterExts(exts, object.Name) {
			u.Path = object.Name
			paths = append(paths,
				fmt.Sprintf("%s\t%s; %s",
					u.String(),
					object.Updated.Local().Format("Jan _2 15:04"),
					bytefmt.Encode(object.Size),
				),
			)
		}
	}
	return paths, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
}
