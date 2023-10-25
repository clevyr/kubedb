package config

import (
	"regexp"
	"strconv"

	"github.com/clevyr/kubedb/internal/consts"
	"github.com/spf13/viper"
)

var (
	namespaceFilterReadOnly  = ".*"
	namespaceFilterReadWrite = "-(dev|stage|test|demo|temp[0-9]+|pr[0-9]+)$"
)

type AccessLevel uint8

const (
	ReadWrite = iota
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

func NewNamespaceRegexp(accessLevel string) NamespaceRegexp {
	v, _ := strconv.ParseUint(accessLevel, 10, 8)
	if AccessLevel(v) == ReadOnly {
		return NamespaceRegexp{re: regexp.MustCompile(namespaceFilterReadOnly)}
	}
	return NamespaceRegexp{re: regexp.MustCompile(namespaceFilterReadWrite)}
}

func init() {
	viper.SetDefault(consts.NamespaceFilterKey, true)
	viper.SetDefault(consts.NamespaceFilterROKey, namespaceFilterReadOnly)
	viper.SetDefault(consts.NamespaceFilterRWKey, namespaceFilterReadWrite)
}
