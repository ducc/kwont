package router

import (
	"context"
	"github.com/ducc/kwɒnt/brokers"
	"github.com/ducc/kwɒnt/protos"
	"github.com/sirupsen/logrus"
	"sync"
)

type BrokerConnections struct {
	sync.RWMutex
	connections    map[string]protos.BrokerServiceClient
	activeSessions map[string]int64
}

func NewBrokerConnections() *BrokerConnections {
	return &BrokerConnections{
		connections:    make(map[string]protos.BrokerServiceClient),
		activeSessions: make(map[string]int64),
	}
}

func (b *BrokerConnections) GetConnections() map[string]protos.BrokerServiceClient {
	b.RLock()
	defer b.RUnlock()

	cpy := make(map[string]protos.BrokerServiceClient)
	for k, v := range b.connections {
		cpy[k] = v
	}
	return cpy
}

func (b *BrokerConnections) GetOrConnect(ctx context.Context, address string) (protos.BrokerServiceClient, error) {
	b.RLock()
	if client, ok := b.connections[address]; ok {
		b.RUnlock()
		return client, nil
	}
	b.RUnlock()

	b.Lock()
	defer b.Unlock()
	client, err := brokers.NewClient(ctx, address)
	if err != nil {
		return nil, err
	}
	logrus.Infof("connected to broker service with address %s", address)
	b.connections[address] = client
	return client, nil
}

func (b *BrokerConnections) SetActiveSessions(address string, sessions int64) {
	b.Lock()
	defer b.Unlock()
	b.activeSessions[address] = sessions
}

func (b *BrokerConnections) FindAddressWithLeastSessions() string {
	b.RLock()
	defer b.RUnlock()

	var min int64
	var addr string
	for address, sessions := range b.activeSessions {
		if addr == "" || sessions < min {
			addr = address
			min = sessions
		}
	}

	return addr
}

func (b *BrokerConnections) RemoveConnection(address string) {
	b.Lock()
	defer b.Unlock()

	delete(b.connections, address)
	delete(b.activeSessions, address)
}
