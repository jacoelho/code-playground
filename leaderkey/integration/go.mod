module integration

go 1.16

replace github.com/jacoelho/leaderkey => ../

require (
	github.com/hashicorp/consul/api v1.8.1
	github.com/jacoelho/leaderkey v0.0.0
	github.com/ory/dockertest/v3 v3.7.0
)
