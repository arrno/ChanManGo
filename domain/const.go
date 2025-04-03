package domain

import "fmt"

const (
	GLOBAL_ADMIN_USER string = "admin"
	MAIN_PORTAL       string = "main"
	MAIN_GROUP        string = "main"
	MAIN_CHANNEL      string = "main"
)

var GLOBAL_ADMIN_ID string = fmt.Sprintf("%s::%s", MAIN_PORTAL, GLOBAL_ADMIN_USER)
