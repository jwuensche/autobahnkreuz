package util

import (
	"context"

	"github.com/gammazero/nexus/client"
	"github.com/gammazero/nexus/wamp"
)

// Ping provides a simple keep-alive function for clients.
func ping(_ context.Context, _ wamp.List, _, _ wamp.Dict) *client.InvokeResult {
	// This function was introduced due to idle-timeouting connections from websockets.
	return &client.InvokeResult{}
}

// RegisterPing registers the ping callback.
func RegisterPing(c *client.Client) error {
	return c.Register("ee.ping", ping, wamp.Dict{})
}
