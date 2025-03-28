package datastore

import (
	"github.com/arrno/chanman/domain"
	"github.com/arrno/chanman/util"
)

// MainPortal -> MainGroup -> MainChannel, AdminClient
func SeedDb(db DatastoreProvider) {
	expect := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	client, err := domain.NewClient("admin", domain.MAIN_GROUP, domain.MAIN_PORTAL, true)
	expect(db.StoreClient(client))
	localSecret := util.NewEnv().Expect("LOCAL_SECRET")
	token, err := domain.NewToken("admin", localSecret)
	expect(err)
	token.Encrypt()
	secret := domain.NewClientSecret(client.Id, token.Encrypted)
	expect(db.StoreClientSecret(&secret))
	portal, err := domain.NewPortal(client.Id, domain.MAIN_PORTAL)
	expect(err)
	portal.Id = domain.MAIN_PORTAL
	expect(db.StorePortal(portal))
	group, err := domain.NewGroup(client.Id, portal.Id, domain.MAIN_PORTAL)
	expect(err)
	group.Id = domain.MAIN_GROUP
	expect(db.StoreGroup(group))
	clientGroup := domain.NewClientGroup(client.Id, group.Id)
	expect(db.StoreClientGroup(clientGroup))
	channel, err := domain.NewChannel(portal.Id, group.Id, domain.MAIN_CHANNEL, 300)
	expect(db.StoreChannel(channel))
}
