package config

import "github.com/spf13/cobra"

func Unmarshal(cmd *cobra.Command, name string, v any) error {
	if !Loaded && cmd != nil {
		if err := Load(cmd); err != nil {
			return err
		}
	}

	if err := K.Unmarshal("", v); err != nil {
		return err
	}

	if name != "" && K.Get(name) != nil {
		if err := K.Unmarshal(name, v); err != nil {
			return err
		}
	}

	return nil
}
