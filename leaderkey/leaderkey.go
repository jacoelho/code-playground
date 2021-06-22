package leaderkey

import (
	"context"
	"time"

	"github.com/hashicorp/consul/api"
)

// https://learn.hashicorp.com/tutorials/consul/application-leader-elections
// https://www.consul.io/docs/dynamic-app-config/sessions#session-design

type membership uint8

const (
	follower membership = iota
	leader
)

type Member struct {
	client                *api.Client
	lock                  *api.Lock
	lockOptions           *api.LockOptions
	membershipChangesChan chan membership
	onLeader              func()
	onFollower            func()
	done                  chan struct{}
}

type MemberOpt func(*Member)

func WithChecks(checks []string) MemberOpt {
	return func(m *Member) {
		m.lockOptions.SessionOpts.Checks = checks
	}
}

func WithNodeChecks(checks []string) MemberOpt {
	return func(m *Member) {
		m.lockOptions.SessionOpts.NodeChecks = checks
	}
}

func WithServiceChecks(checks []api.ServiceCheck) MemberOpt {
	return func(m *Member) {
		m.lockOptions.SessionOpts.ServiceChecks = checks
	}
}

func WithSessionName(name string) MemberOpt {
	return func(m *Member) {
		m.lockOptions.SessionOpts.Name = name
	}
}

func WithSessionTTL(ttl time.Duration) MemberOpt {
	return func(m *Member) {
		m.lockOptions.SessionOpts.TTL = ttl.String()
	}
}

func WithLockDelay(delay time.Duration) MemberOpt {
	return func(m *Member) {
		m.lockOptions.LockDelay = delay
	}
}

func WithOnLeaderFunc(f func()) MemberOpt {
	return func(m *Member) {
		m.onLeader = f
	}
}

func WithOnFollowerFunc(f func()) MemberOpt {
	return func(m *Member) {
		m.onFollower = f
	}
}

func New(cfg *api.Config, key string, opts ...MemberOpt) (*Member, error) {
	if cfg == nil {
		cfg = api.DefaultConfig()
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	m := &Member{
		client:                client,
		onLeader:              func() {},
		onFollower:            func() {},
		membershipChangesChan: make(chan membership),
		done:                  make(chan struct{}, 1),
		lockOptions: &api.LockOptions{
			Key: key,
			SessionOpts: &api.SessionEntry{
				Behavior: api.SessionBehaviorRelease,
			},
			SessionName: api.DefaultLockSessionName,
			SessionTTL:  api.DefaultLockSessionTTL,
			LockDelay:   api.DefaultLockWaitTime,
		},
	}

	for _, opt := range opts {
		opt(m)
	}

	go func() {
		for {
			select {
			case c := <-m.membershipChangesChan:
				switch c {
				case leader:
					m.onLeader()
				default:
					m.onFollower()
				}
			case <-m.done:
				return
			}
		}
	}()

	return m, nil
}

func (c *Member) Run(ctx context.Context) error {
	var err error
	c.lock, err = c.client.LockOpts(c.lockOptions)
	if err != nil {
		return err
	}

loop:
	for {
		lostCh, err := c.lock.Lock(ctx.Done())
		if err != nil {
			continue
		}

		if lostCh != nil {
			c.membershipChangesChan <- leader
		}

		select {
		case <-lostCh:
			c.membershipChangesChan <- follower

		case <-ctx.Done():
			break loop

		case <-c.done:
			break loop
		}
	}

	close(c.done)

	return nil
}

func (c *Member) Stop() {
	close(c.done)
}
