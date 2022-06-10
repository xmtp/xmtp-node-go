package server

import "testing"

func TestServerStartStop(t *testing.T) {
	var options Options
	// requires the docker-compose setup to be running
	options.Store.DbConnectionString = "postgres://postgres:xmtp@localhost:5432/postgres?sslmode=disable"
	options.NodeKey = "8a30dcb604b0b53627a5adc054dbf434b446628d4bd1eccc681d223f0550ce67"
	options.Address = "localhost"
	options.Relay.Disable = true
	server := New(options)
	server.Shutdown()
}
