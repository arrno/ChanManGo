package datastore

import (
	"github.com/arrno/chanman/domain"
)

// MainPortal -> MainGroup -> MainChannel, AdminClient
func SeedDb(db RepoProvider) {
	// if _, ok := db.(*LocalDB); !ok { // only seed local
	// 	return
	// }
	expect := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	clients, _ := db.GetClients(&domain.Pagination[*domain.Client]{Page: 1, Count: 10}, nil)
	if len(clients) > 0 {
		return
	}
	groupId := domain.MAIN_GROUP
	client, err := domain.NewClient(domain.GLOBAL_ADMIN, domain.GLOBAL_ADMIN, groupId, true)
	expect(db.StoreClient(client))

	token := domain.NewToken("password", client.Id).Hash()

	secret := domain.NewClientSecretFromToken(*token)
	expect(db.StoreClientSecret(&secret))

	expect(db.StoreClientPermission(domain.NewReadWrite(client.Id)))

	portal, err := domain.NewPortal(domain.MAIN_PORTAL, client.Id, domain.MAIN_PORTAL)
	expect(err)

	portal.Id = domain.MAIN_PORTAL
	expect(db.StorePortal(portal))

	group, err := domain.NewGroup(domain.MAIN_GROUP, client.Id, portal.Id, domain.MAIN_GROUP)
	expect(err)

	expect(db.StoreGroup(group))

	clientGroup := domain.NewClientGroup(client.Id, group.Id)
	expect(db.StoreClientGroup(clientGroup))

	channel, err := domain.NewChannel(domain.MAIN_CHANNEL, portal.Id, group.Id, domain.MAIN_CHANNEL, 60)
	expect(db.StoreChannel(channel))
}
