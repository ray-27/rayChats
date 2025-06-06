package db

import "os"

var (
	Store *ValkeyChatStore
)

func DB_init() {
	valkey_endpoint := os.Getenv("VALKEY_ENDPOINT")
	valkey_password := os.Getenv("VALKEY_PASSWORD")
	Store = NewValkeyChatStore(valkey_endpoint, valkey_password, 1) //this `1` is for the room information partation
}
