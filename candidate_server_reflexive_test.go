// SPDX-FileCopyrightText: 2023 The Pion community <https://pion.ly>
// SPDX-License-Identifier: MIT

//go:build !js
// +build !js

package ice

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/theodorsm/stun/v2"
	"github.com/pion/transport/v3/test"
	"github.com/theodorsm/turn/v3"
	"github.com/stretchr/testify/require"
)

func TestServerReflexiveOnlyConnection(t *testing.T) {
	defer test.CheckRoutines(t)()

	// Limit runtime in case of deadlocks
	defer test.TimeOut(time.Second * 30).Stop()

	serverPort := randomPort(t)
	serverListener, err := net.ListenPacket("udp4", "127.0.0.1:"+strconv.Itoa(serverPort))
	require.NoError(t, err)

	server, err := turn.NewServer(turn.ServerConfig{
		Realm:       "pion.ly",
		AuthHandler: optimisticAuthHandler,
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn:            serverListener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorNone{Address: "127.0.0.1"},
			},
		},
	})
	require.NoError(t, err)

	cfg := &AgentConfig{
		NetworkTypes: []NetworkType{NetworkTypeUDP4},
		Urls: []*stun.URI{
			{
				Scheme: SchemeTypeSTUN,
				Host:   "127.0.0.1",
				Port:   serverPort,
			},
		},
		CandidateTypes: []CandidateType{CandidateTypeServerReflexive},
	}

	aAgent, err := NewAgent(cfg)
	require.NoError(t, err)

	aNotifier, aConnected := onConnected()
	require.NoError(t, aAgent.OnConnectionStateChange(aNotifier))

	bAgent, err := NewAgent(cfg)
	require.NoError(t, err)

	bNotifier, bConnected := onConnected()
	require.NoError(t, bAgent.OnConnectionStateChange(bNotifier))

	connect(aAgent, bAgent)
	<-aConnected
	<-bConnected

	require.NoError(t, aAgent.Close())
	require.NoError(t, bAgent.Close())
	require.NoError(t, server.Close())
}
