package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/jacoelho/leaderkey"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func TestFoo(t *testing.T) {
	cfg := api.DefaultConfig()
	cfg.Address = newTestConsul(t)

	client, err := api.NewClient(cfg)
	if err != nil {
		t.Fatalf("expected client %v", err)
	}

	leader, err := leaderkey.New(client, "abc", 15*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)

	foo := make(chan struct{}, 1)
	errChan := make(chan error, 1)
	go func() {
		err = leader.TryLockContinuously(ctx, func() {
			fmt.Println("inside")
			time.Sleep(5 * time.Second)
			foo <- struct{}{}
		}, nil)
		if err != nil {
			errChan <- err
		}
	}()

	select {
	case <-foo:

	case err := <-errChan:
		t.Error(err)
	}

	cancel()

}

func newTestConsul(t *testing.T) string {
	t.Helper()

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}

	consul, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "consul",
		Tag:        "1.9",
		Cmd:        []string{"agent", "-server", "-node=server-1", "-bootstrap-expect=1", "-client=0.0.0.0"},
	}, func(c *docker.HostConfig) {
		//c.AutoRemove = true
		c.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		t.Fatalf("Could not start resource: %v", err)
	}

	if err := consul.Expire(60); err != nil {
		t.Fatalf("failed to expire database container: %v", err)
	}

	t.Cleanup(func() {
		// if err := pool.Purge(consul); err != nil {
		// 	t.Fatalf("failed to cleanup consul container: %v", err)
		// }
	})

	return consul.GetHostPort("8500/tcp")
}
