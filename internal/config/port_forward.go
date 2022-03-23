package config

type PortForward struct {
	Global
	Addresses []string
	LocalPort uint16
}
