package storage

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"

	"gabe565.com/utils/bytefmt"
	"github.com/clevyr/kubedb/internal/util"
	"github.com/spf13/cobra"
)

func CompleteBuckets(u *url.URL) ([]string, cobra.ShellCompDirective) {
	u.Path = "/"

	client, err := NewClient(context.Background(), u.String())
	if err != nil {
		slog.Error("Failed to create storage client", "error", err)
		return nil, cobra.ShellCompDirectiveError
	}

	names := make([]string, 0, 16)
	for bucket, err := range client.ListBuckets(context.Background()) {
		if err != nil {
			slog.Error("Failed to list storage buckets", "error", err)
			return nil, cobra.ShellCompDirectiveError
		}

		u.Host = bucket.Name
		names = append(names, u.String())
	}
	return names, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
}

func CompleteObjects(u *url.URL, exts []string, dirOnly bool) ([]string, cobra.ShellCompDirective) {
	client, err := NewClient(context.Background(), u.String())
	if err != nil {
		slog.Error("Failed to create storage client", "error", err)
		return nil, cobra.ShellCompDirectiveError
	}

	paths := make([]string, 0, 16)
	for object, err := range client.ListObjects(context.Background(), u.String()) {
		if err != nil {
			slog.Error("Failed to list storage bucket objects", "error", err)
			return nil, cobra.ShellCompDirectiveError
		}

		u.Path = object.Name
		if object.IsDir {
			paths = append(paths, u.String())
		} else if !dirOnly && util.FilterExts(exts, object.Name) {
			paths = append(paths,
				fmt.Sprintf("%s\t%s; %s",
					u.String(),
					object.LastModified.Local().Format("Jan _2 15:04"),
					bytefmt.Encode(object.Size),
				),
			)
		}
	}
	return paths, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
}
