package models

type ProxySource interface {
	Create(opts ...interface{})
	GetProxylist() string
	GetProxyStrings(data string) []string
	ConvertStrings(data []string) []*Proxy
}
