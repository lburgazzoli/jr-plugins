//go:build plugin_pulsar
// +build plugin_pulsar

// Copyright © 2024 JR team
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package pulsar

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/jrnd-io/jr-plugins/internal/plugin"
	"github.com/jrnd-io/jrv2/pkg/jrpc"
)

var _ plugin.Plugin = (*Plugin)(nil)

const (
	Name = "pulsar"
)

func init() {
	plugin.RegisterPlugin(Name, &Plugin{})
}

type Plugin struct {
	client   pulsar.Client
	producer pulsar.Producer
}

func (p *Plugin) Init(ctx context.Context, cfgBytes []byte) error {
	config := Config{}
	if err := json.Unmarshal(cfgBytes, &config); err != nil {
		return fmt.Errorf("failed to unmarshal Pulsar plugin config: %w", err)
	}

	if err := p.InitializeFromConfig(ctx, config); err != nil {
		return fmt.Errorf("failed to initialize Pulsar plugin config: %w", err)
	}

	return nil
}

func (p *Plugin) InitializeFromConfig(_ context.Context, config Config) error {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               config.URL,
		ConnectionTimeout: config.ConnectionTimeout,
		OperationTimeout:  config.OperationTimeout,
		KeepAliveInterval: config.KeepAliveInterval,
	})

	if err != nil {
		return fmt.Errorf("failed to create pulsar client: %w", err)
	}

	p.client = client

	producer, err := p.client.CreateProducer(pulsar.ProducerOptions{
		Topic: config.Topic,
	})
	if err != nil {
		return fmt.Errorf("failed to create pulsar producer: %w", err)
	}

	p.producer = producer

	return nil
}

func (p *Plugin) Produce(k []byte, v []byte, h map[string]string) (*jrpc.ProduceResponse, error) {
	_, err := p.producer.Send(context.Background(), &pulsar.ProducerMessage{
		Key:        string(k),
		Payload:    v,
		Properties: h,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to send message to Pulsar: %w", err)
	}

	return nil, nil
}

func (p *Plugin) Close(_ context.Context) error {
	if p.producer != nil {
		p.producer.Close()
	}
	if p.client != nil {
		p.client.Close()
	}

	return nil
}
