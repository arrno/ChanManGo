package service

type SystemProvider interface {
	ClientProvider
	GroupProvider
	PortalProvider
}

func NewSystem() SystemProvider {
	return nil
}
