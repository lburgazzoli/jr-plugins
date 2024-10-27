package pulsar_test

import (
	"context"
	"fmt"
	pgo "github.com/apache/pulsar-client-go/pulsar"
	"github.com/apache/pulsar-client-go/pulsaradmin"
	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/admin"
	"github.com/docker/go-connections/nat"
	"github.com/rs/xid"
	"io"
	"strings"

	tc "github.com/testcontainers/testcontainers-go"
	tcwait "github.com/testcontainers/testcontainers-go/wait"
)

const (
	PulsarImage                       = "apachepulsar/pulsar:3.3.0"
	PulsarBrokerPort                  = "6650/tcp"
	PulsarBrokerHTTPPort              = "8080/tcp"
	PulsarBrokerAdminClustersEndpoint = "/admin/v2/clusters"
	PulsarBrokerCMD                   = "/pulsar/bin/apply-config-from-env.py /pulsar/conf/standalone.conf && bin/pulsar standalone --no-functions-worker -nss"
)

type PulsarContainer struct {
	tc.Container
}

func (c *PulsarContainer) BrokerURL(ctx context.Context) (string, error) {
	return c.resolveURL(ctx, PulsarBrokerPort)
}

func (c *PulsarContainer) HTTPServiceURL(ctx context.Context) (string, error) {
	return c.resolveURL(ctx, PulsarBrokerHTTPPort)
}

func (c *PulsarContainer) resolveURL(ctx context.Context, port nat.Port) (string, error) {
	provider, err := tc.NewDockerProvider()
	if err != nil {
		return "", err
	}
	defer provider.Close()

	host, err := provider.DaemonHost(ctx)
	if err != nil {
		return "", err
	}

	mp, err := c.MappedPort(ctx, port)
	if err != nil {
		return "", err
	}

	proto := "pulsar"
	if port == PulsarBrokerHTTPPort {
		proto = "http"
	}

	return fmt.Sprintf("%s://%s:%v", proto, host, mp.Int()), nil
}

func (c *PulsarContainer) Admin(ctx context.Context) (admin.Client, error) {
	url, err := c.HTTPServiceURL(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get URL for pulsar admin: %v", err)
	}

	cfg := &pulsaradmin.Config{
		WebServiceURL: url,
	}

	adminClient, err := pulsaradmin.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return adminClient, nil
}

func (c *PulsarContainer) Subscribe(ctx context.Context, topic string) (pgo.Consumer, error) {
	url, err := c.BrokerURL(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get URL for pulsar admin: %v", err)
	}

	client, err := pgo.NewClient(pgo.ClientOptions{
		URL: url,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create pulsar client: %w", err)
	}

	consumer, err := client.Subscribe(pgo.ConsumerOptions{
		Topic:                       topic,
		SubscriptionName:            xid.New().String(),
		Type:                        pgo.AUTO,
		SubscriptionInitialPosition: pgo.SubscriptionPositionEarliest,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create pulsar consumer: %w", err)
	}

	return consumer, nil
}

func RunPulsarContainer(ctx context.Context, opts ...tc.ContainerCustomizer) (*PulsarContainer, error) {
	req := tc.ContainerRequest{
		Image:        PulsarImage,
		Env:          map[string]string{},
		ExposedPorts: []string{PulsarBrokerPort, PulsarBrokerHTTPPort},
		Cmd:          []string{"/bin/bash", "-c", PulsarBrokerCMD},
		WaitingFor: tcwait.ForHTTP(PulsarBrokerAdminClustersEndpoint).
			WithPort(PulsarBrokerHTTPPort).
			WithResponseMatcher(func(body io.Reader) bool {
				c, err := io.ReadAll(body)
				if err != nil {
					return false
				}

				return strings.Contains(string(c), "standalone")
			}),
	}

	genericContainerReq := tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	container, err := tc.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, fmt.Errorf("generic container: %w", err)
	}

	var c *PulsarContainer
	if container != nil {
		c = &PulsarContainer{
			Container: container,
		}
	}

	return c, nil
}
