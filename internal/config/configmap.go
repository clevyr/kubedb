package config

type Permission struct {
	Read  bool
	Write bool
}

type ConfigMap struct {
	Permission Permission
}
