package config

import (
	"os"
	"strings"

	"gabe565.com/utils/must"
	"github.com/clevyr/kubedb/internal/config/conftypes"
	"github.com/clevyr/kubedb/internal/config/path"
	"github.com/clevyr/kubedb/internal/consts"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals
var (
	K            = koanf.New(".")
	Global       = &conftypes.Global{}
	Loaded       bool
	IsCompletion bool
)

const EnvPrefix = "KUBEDB_"

func Load(cmd *cobra.Command) error {
	if err := K.Load(confmap.Provider(map[string]any{
		"log-mask": true,
		"namespace-colors": map[string]string{
			"[-_]pro?d(uction)?([-_]|$)": "1",
		},
	}, "."), nil); err != nil {
		return err
	}

	// Find config file
	cfgFile := must.Must2(cmd.Flags().GetString(consts.FlagConfig))
	if cfgFile == "" {
		var err error
		cfgFile, err = path.GetConfigFile()
		if err != nil {
			return err
		}
	}
	if strings.Contains(cfgFile, "$HOME") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		cfgFile = strings.Replace(cfgFile, "$HOME", home, 1)
	}

	// Load config file
	cfgContents, err := os.ReadFile(cfgFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if len(cfgContents) != 0 {
		if err := K.Load(rawbytes.Provider(cfgContents), yaml.Parser()); err != nil {
			return err
		}
	}

	// Load envs
	if err := K.Load(env.Provider(EnvPrefix, ".", func(s string) string {
		s = strings.TrimPrefix(s, EnvPrefix)
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, "_", "-")
		return s
	}), nil); err != nil {
		return err
	}
	if os.Getenv(EnvPrefix+"KUBECONFIG") == "" {
		if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
			must.Must(K.Set(consts.FlagKubeConfig, kubeconfig))
		}
	}

	// Load flags
	if err := K.Load(posflag.ProviderWithValue(cmd.Flags(), ".", K, func(key string, value string) (string, any) {
		f := cmd.Flags().Lookup(key)
		switch f.Value.Type() {
		case "stringSlice":
			return key, must.Must2(cmd.Flags().GetStringSlice(key))
		case "stringToString":
			return key, must.Must2(cmd.Flags().GetStringToString(key))
		default:
			return key, value
		}
	}), nil); err != nil {
		return err
	}

	kubeconfig := K.String(consts.FlagKubeConfig)
	if kubeconfig == "$HOME" || strings.HasPrefix(kubeconfig, "$HOME"+string(os.PathSeparator)) {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		kubeconfig = home + kubeconfig[5:]
		if err := K.Set(consts.FlagKubeConfig, kubeconfig); err != nil {
			panic(err)
		}
	}

	Loaded = true
	return Unmarshal(nil, "", Global)
}
