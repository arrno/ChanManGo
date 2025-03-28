package datastore

import (
	"github.com/arrno/chanman/domain"
	"github.com/arrno/chanman/util"
)

// MainPortal -> MainGroup -> MainChannel, AdminClient
func SeedDb(db DatastoreProvider) {
	client := domain.NewClient("admin", domain.MAIN_GROUP, domain.MAIN_PORTAL, true)
	err := db.StoreClient(client)
	if err != nil {
		panic(err)
	}
	localSecret := util.NewEnv().Expect("LOCAL_SECRET")
	token, err := domain.NewToken("admin", localSecret)
	if err != nil {
		panic(err)
	}
	token.Encrypt()
	secret := domain.NewClientSecret(client.Id, token.Encrypted)
	err = db.StoreClientSecret(&secret)
	if err != nil {
		panic(err)
	}
	portal := domain.NewPortal(client.Id, domain.MAIN_PORTAL)
	portal.Id = domain.MAIN_PORTAL
	err = db.StorePortal(portal)
	if err != nil {
		panic(err)
	}
	group := domain.NewGroup(client.Id, portal.Id, domain.MAIN_PORTAL)
	group.Id = domain.MAIN_GROUP
	err = db.StoreGroup(group)
	if err != nil {
		panic(err)
	}
	clientGroup := domain.NewClientGroup(client.Id, group.Id)
	err = db.StoreClientGroup(clientGroup)
	if err != nil {
		panic(err)
	}
	channel := domain.NewChannel(portal.Id, group.Id, domain.MAIN_CHANNEL, 300)
	err = db.StoreChannel(channel)
	if err != nil {
		panic(err)
	}
}
