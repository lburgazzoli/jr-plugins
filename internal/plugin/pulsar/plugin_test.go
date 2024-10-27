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

package pulsar_test

import (
	"bytes"
	"context"
	pgo "github.com/apache/pulsar-client-go/pulsar"
	"github.com/apache/pulsar-client-go/pulsaradmin/pkg/utils"
	"github.com/jrnd-io/jr-plugins/internal/plugin/pulsar"
	"github.com/rs/xid"
	tc "github.com/testcontainers/testcontainers-go"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

type TestLogConsumer struct {
	t *testing.T
}

// Accept prints the log to stdout
func (lc *TestLogConsumer) Accept(l tc.Log) {
	c := string(l.Content)
	c = strings.TrimSpace(c)

	lc.t.Logf("[pulsar] " + c)
}

func TestPulsarPlugin(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)

	topic := xid.New().String()
	logger := tc.TestLogger(t)

	pulsarContainer, err := RunPulsarContainer(
		ctx,
		tc.WithLogger(logger),
		//tc.WithLogConsumers(&TestLogConsumer{t: t}),
	)

	t.Cleanup(func() {
		_ = tc.TerminateContainer(pulsarContainer)
	})

	g.Expect(err).ToNot(HaveOccurred())

	brokerURL, err := pulsarContainer.BrokerURL(ctx)
	g.Expect(err).ToNot(HaveOccurred())

	ca, err := pulsarContainer.Admin(ctx)
	g.Expect(err).ToNot(HaveOccurred())

	tn, err := utils.GetTopicName(topic)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(tn).ToNot(BeNil())

	err = ca.Topics().Create(*tn, 1)
	g.Expect(err).ToNot(HaveOccurred())

	cfg := pulsar.Config{
		URL:   brokerURL,
		Topic: topic,
	}

	p := &pulsar.Plugin{}
	err = p.InitializeFromConfig(ctx, cfg)

	t.Cleanup(func() {
		_ = p.Close(context.Background())
	})

	g.Expect(err).ToNot(HaveOccurred())

	_, err = p.Produce([]byte("somekey"), []byte("someval"), nil)
	g.Expect(err).ToNot(HaveOccurred())

	s, err := pulsarContainer.Subscribe(ctx, topic)
	g.Expect(err).ToNot(HaveOccurred())

	g.Eventually(s.Receive, ctx).Should(Satisfy(func(in pgo.Message) bool {
		return in.Key() == "somekey" && bytes.Equal(in.Payload(), []byte("someval"))
	}))
}
