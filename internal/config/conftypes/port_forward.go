package conftypes

type PortForward struct {
	*Global    `koanf:"-"`
	Address    []string `koanf:"address"`
	ListenPort uint16   `koanf:"listen-port"`
}
