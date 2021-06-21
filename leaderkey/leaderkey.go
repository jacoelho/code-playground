package leaderkey

import (
	"context"
	"time"

	"github.com/hashicorp/consul/api"
)

// https://learn.hashicorp.com/tutorials/consul/application-leader-elections

type LockClient struct {
	client *api.Client
	key    string
	ttl    time.Duration
}

func New(client *api.Client, key string, ttl time.Duration) (*LockClient, error) {
	if client == nil {
		var err error
		client, err = api.NewClient(api.DefaultConfig())
		if err != nil {
			return nil, err
		}
	}

	return &LockClient{
		client: client,
		key:    key,
		ttl:    ttl,
	}, nil
}

func (c *LockClient) TryLockContinuously(ctx context.Context, onLeader func(), onFollower func()) error {
	if onLeader == nil {
		onLeader = func() {}
	}

	if onFollower == nil {
		onFollower = func() {}
	}

	se := &api.SessionEntry{
		Name:     api.DefaultLockSessionName,
		TTL:      c.ttl.String(),
		Behavior: api.SessionBehaviorDelete,
	}

	id, _, err := c.client.Session().CreateNoChecks(se, nil)
	if err != nil {
		return err
	}

	opts := &api.LockOptions{
		Key:              c.key,
		Session:          id,
		SessionName:      se.Name,
		SessionTTL:       se.TTL,
		MonitorRetries:   5,
		MonitorRetryTime: 2 * time.Second,
	}

	lock, err := c.client.LockOpts(opts)
	if err != nil {
		return err
	}

	for {
		lostCh, err := lock.Lock(ctx.Done())
		if err != nil {
			continue
		}

		onLeader()

		select {
		case <-lostCh:
			onFollower()

		case <-ctx.Done():
			return nil
		}
	}
}
