package service

type AnalyticsProvider interface {
	ChannelRates()  // all / by topic / sort
	ChannelErrors() // all / by topic / sort
	//
	ClientRates()  // all / by topic / by channel / sort
	ClientErrors() // all / by topic / by channel / sort
	//
	GroupOverview()
	//
	PortalOverview()
}
