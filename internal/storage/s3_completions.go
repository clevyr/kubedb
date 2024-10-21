package storage

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"

	"github.com/clevyr/kubedb/internal/util"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

func CompleteBucketsS3(u *url.URL) ([]string, cobra.ShellCompDirective) {
	u.Path = "/"

	var names []string
	for output, err := range ListBucketsS3(context.Background(), nil) {
		if err != nil {
			slog.Error("Failed to list S3 buckets", "error", err)
			return nil, cobra.ShellCompDirectiveError
		}

		names = slices.Grow(names, len(output.Buckets))

		for _, bucket := range output.Buckets {
			u.Host = *bucket.Name
			names = append(names, u.String())
		}
	}
	return names, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
}

func CompleteObjectsS3(u *url.URL, exts []string, dirOnly bool) ([]string, cobra.ShellCompDirective) {
	var paths []string
	for output, err := range ListObjectsS3(context.Background(), u.String()) {
		if err != nil {
			slog.Error("Failed to list S3 objects", "error", err)
			return nil, cobra.ShellCompDirectiveError
		}

		paths = slices.Grow(paths, len(output.CommonPrefixes)+len(output.Contents))

		for _, prefix := range output.CommonPrefixes {
			u.Path = *prefix.Prefix
			paths = append(paths, u.String())
		}

		if !dirOnly {
			for _, object := range output.Contents {
				if !strings.HasSuffix(*object.Key, "/") && !util.FilterExts(exts, *object.Key) {
					continue
				}

				u.Path = *object.Key
				paths = append(paths,
					fmt.Sprintf("%s\t%s; %s",
						u.String(),
						object.LastModified.Local().Format("Jan _2 15:04"), //nolint:gosmopolitan
						humanize.IBytes(uint64(*object.Size)),              //nolint:gosec
					),
				)
			}
		}
	}
	return paths, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
}
