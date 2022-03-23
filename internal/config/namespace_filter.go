package config

import (
	"github.com/spf13/viper"
	"regexp"
	"strconv"
)

var (
	namespaceFilterReadOnly  = ".*"
	namespaceFilterReadWrite = "-(dev|stage|test|demo|pr[0-9]+)$"
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
	if !viper.GetBool("namespace.filter") {
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
	viper.SetDefault("namespace.filter", true)
	viper.SetDefault("namespace.ro", namespaceFilterReadOnly)
	viper.SetDefault("namespace.rw", namespaceFilterReadWrite)
}
