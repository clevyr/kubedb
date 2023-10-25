package namespace_filter

import (
	"context"
	"regexp"

	"github.com/clevyr/kubedb/internal/consts"
	"github.com/spf13/viper"
)

var (
	namespaceFilterReadOnly  = ".*"
	namespaceFilterReadWrite = "-(dev|stage|test|demo|temp[0-9]+|pr[0-9]+)$"
)

type AccessLevel uint8

const (
	ReadWrite AccessLevel = iota
	ReadOnly
)

type NamespaceRegexp struct {
	re *regexp.Regexp
}

func (level NamespaceRegexp) Match(namespace string) bool {
	if !viper.GetBool(consts.NamespaceFilterKey) {
		return true
	}
	return level.re.MatchString(namespace)
}

func NewFromContext(ctx context.Context) NamespaceRegexp {
	accessLevel, _ := LevelFromContext(ctx)

	switch accessLevel {
	case ReadOnly:
		return NamespaceRegexp{re: regexp.MustCompile(namespaceFilterReadOnly)}
	default:
		return NamespaceRegexp{re: regexp.MustCompile(namespaceFilterReadWrite)}
	}
}

func init() {
	viper.SetDefault(consts.NamespaceFilterKey, true)
	viper.SetDefault(consts.NamespaceFilterROKey, namespaceFilterReadOnly)
	viper.SetDefault(consts.NamespaceFilterRWKey, namespaceFilterReadWrite)
}

type contextKey string

var accessLevelKey = contextKey("accessLevel")

func NewContext(ctx context.Context, level AccessLevel) context.Context {
	return context.WithValue(ctx, accessLevelKey, level)
}

func LevelFromContext(ctx context.Context) (AccessLevel, bool) {
	filter, ok := ctx.Value(accessLevelKey).(AccessLevel)
	return filter, ok
}
