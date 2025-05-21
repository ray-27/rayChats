package db

var (
	Store *ValkeyChatStore
)

func DB_init() {
	Store = NewValkeyChatStore("localhost:6379", "", 1)
}
