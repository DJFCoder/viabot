module github.com/viabot/customer-a

go 1.22

require (
	viabot.stream/sdk v0.0.0
	github.com/mattn/go-sqlite3 v1.14.44 // indirect
)

replace viabot.stream/sdk => ../../core
