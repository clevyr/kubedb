package storage

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/clevyr/kubedb/internal/util"
	"github.com/dustin/go-humanize"
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

	var names []string //nolint:prealloc
	for bucket, err := range ListBucketsGCS(context.Background(), projectID) {
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		u.Host = bucket.Name
		names = append(names, u.String())
	}
	return names, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
}

func CompleteObjectsGCS(u *url.URL, exts []string, dirOnly bool) ([]string, cobra.ShellCompDirective) {
	var paths []string
	for object, err := range ListObjectsGCS(context.Background(), u.String()) {
		if err != nil {
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
					object.Updated.Local().Format("Jan _2 15:04"), //nolint:gosmopolitan
					humanize.IBytes(uint64(object.Size)),          //nolint:gosec
				),
			)
		}
	}
	return paths, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
}
