package router

import (
	"context"
	"github.com/ducc/kwɒnt/brokers"
	"github.com/ducc/kwɒnt/protos"
	"sync"
)

type BrokerConnections struct {
	sync.RWMutex
	connections map[string]protos.BrokerServiceClient
}

func NewBrokerConnections() *BrokerConnections {
	return &BrokerConnections{
		connections: make(map[string]protos.BrokerServiceClient),
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

	b.connections[address] = client
	return client, nil
}
