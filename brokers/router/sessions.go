package router

import (
	"context"
	"errors"
	"fmt"
	"github.com/ducc/kwɒnt/protos"
	"github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
	"strings"
)

type SessionFinder struct {
	client *redis.Client
}

func NewSessionFinder(ctx context.Context, address string) (*SessionFinder, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       0,
		PoolSize: 20,
	})

	if _, err := client.WithContext(ctx).Ping().Result(); err != nil {
		return nil, err
	}

	return &SessionFinder{
		client: client,
	}, nil
}

func (r *SessionFinder) getBytes(ctx context.Context, key string) ([]byte, error) {
	return r.client.WithContext(ctx).Get(key).Bytes()
}

func (r *SessionFinder) setBytes(ctx context.Context, key string, val []byte) error {
	return r.client.WithContext(ctx).Set(key, val, 0).Err() // todo expiration
}

var ErrSessionNotFound = errors.New("session not found")

func (r *SessionFinder) getSessionInfo(ctx context.Context, broker protos.Broker_Name, sessionID string) (*protos.SessionInfo, error) {
	bytes, err := r.getBytes(ctx, fmt.Sprintf("brokers:%s:sessions:%s", strings.ToLower(broker.String()), sessionID))
	if err != nil {
		return nil, err
	}

	if len(bytes) == 0 {
		return nil, ErrSessionNotFound
	}

	var sessionInfo protos.SessionInfo
	if err := proto.Unmarshal(bytes, &sessionInfo); err != nil {
		return nil, err
	}

	return &sessionInfo, nil
}

func (r *SessionFinder) setSessionInfo(ctx context.Context, sessionInfo *protos.SessionInfo) error {
	data, err := proto.Marshal(sessionInfo)
	if err != nil {
		return err
	}

	return r.setBytes(ctx, fmt.Sprintf("brokers:%s:sessions:%s", strings.ToLower(sessionInfo.Broker.String()), sessionInfo.SessionId), data)
}
