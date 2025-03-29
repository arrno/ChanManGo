package controller

import (
	"net/http"

	"github.com/arrno/chanman/broker"
	"github.com/arrno/chanman/datastore"
	"github.com/arrno/chanman/limiter"
	"github.com/arrno/chanman/service"
	ioService "github.com/arrno/chanman/service/io"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections, modify this in production for better security
	},
}

type Controller struct {
	db          datastore.DatastoreProvider
	chanMan     service.ManagerProvider
	system      ioService.SystemProvider
	rateLimiter limiter.LimitProvider
}

type MessageWrapper struct {
	Topic   string `json:"topic"`
	Payload string `json:"payload"`
}

func NewController() *Controller {
	db := datastore.NewInMemoryDB()
	datastore.SeedDb(db)
	broker := broker.NewInMemoryBroker()
	limiter := limiter.NewLocalLimiter()
	return &Controller{
		db:          db,
		chanMan:     service.NewManager(db, broker),
		system:      ioService.NewSystem(db),
		rateLimiter: limiter,
	}
}

func parseHeaders(h http.Header) (clientId, groupId, portalId, channelName, read, write string) {
	clientId = h.Get("clientId")
	groupId = h.Get("groupId")
	portalId = h.Get("portalId")
	channelName = h.Get("channel")
	read = h.Get("read")
	write = h.Get("write")
	return
}

func (c *Controller) GetChannel(w http.ResponseWriter, r *http.Request)     {}
func (c *Controller) CreateChannel(w http.ResponseWriter, r *http.Request)  {}
func (c *Controller) UpdateChannel(w http.ResponseWriter, r *http.Request)  {}
func (c *Controller) ArchiveChannel(w http.ResponseWriter, r *http.Request) {}

func (c *Controller) GetGroup(w http.ResponseWriter, r *http.Request)     {}
func (c *Controller) CreateGroup(w http.ResponseWriter, r *http.Request)  {}
func (c *Controller) UpdateGroup(w http.ResponseWriter, r *http.Request)  {}
func (c *Controller) ArchiveGroup(w http.ResponseWriter, r *http.Request) {}

func (c *Controller) AddClientToGroup(w http.ResponseWriter, r *http.Request)       {}
func (c *Controller) RemoveClientFromGroup(w http.ResponseWriter, r *http.Request)  {}
func (c *Controller) AddChannelToGroup(w http.ResponseWriter, r *http.Request)      {}
func (c *Controller) RemoveChannelFromGroup(w http.ResponseWriter, r *http.Request) {}

func (c *Controller) CreatePortal(w http.ResponseWriter, r *http.Request)  {}
func (c *Controller) UpdatePortal(w http.ResponseWriter, r *http.Request)  {}
func (c *Controller) ArchivePortal(w http.ResponseWriter, r *http.Request) {}
