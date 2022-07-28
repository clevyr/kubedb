package config

type PortForward struct {
	Global    `mapstructure:",squash"`
	Addresses []string
	LocalPort uint16
}
