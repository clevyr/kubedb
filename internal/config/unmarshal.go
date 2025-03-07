package config

func Unmarshal(name string, v any) error {
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
