//go:build http
// +build http

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

package http

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jrnd-io/jr-plugins/internal/plugin"
	"github.com/jrnd-io/jrv2/pkg/jrpc"
)

const (
	Name = "http"
)

func init() {
	plugin.RegisterPlugin(Name, &Plugin{})
}

type Plugin struct {
	configuration Config

	certificate tls.Certificate
	client      *resty.Client
	cookiejar   http.CookieJar
}

func (p *Plugin) Init(_ context.Context, cfgBytes []byte) error {

	config := Config{}
	if err := json.Unmarshal(cfgBytes, &config); err != nil {
		return err
	}

	return p.InitializeFromConfig(config)
}

func (p *Plugin) InitializeFromConfig(config Config) error {

	var err error
	p.configuration = config
	if p.configuration.Endpoint.Timeout == "" {
		p.configuration.Endpoint.timeout = time.Second * 10
	} else {
		p.configuration.Endpoint.timeout, err = time.ParseDuration(p.configuration.Endpoint.Timeout)
		if err != nil {
			return err
		}

	}

	if p.configuration.ErrorHandling.ExpectStatusCode == 0 {
		p.configuration.ErrorHandling.ExpectStatusCode = 200
	}

	if p.configuration.TLS.CertFile != "" && p.configuration.TLS.KeyFile == "" {
		return fmt.Errorf("CertFile is set but KeyFile is not")
	}
	if p.configuration.TLS.CertFile == "" && p.configuration.TLS.KeyFile != "" {
		return fmt.Errorf("KeyFile is set but CertFile is not")
	}

	certificates := make([]tls.Certificate, 0)
	if p.configuration.TLS.CertFile != "" {
		p.certificate, err = tls.LoadX509KeyPair(p.configuration.TLS.CertFile, p.configuration.TLS.KeyFile)
		if err != nil {
			return err
		}
		certificates = append(certificates, p.certificate)
	}

	if p.configuration.Session.UseCookieJar {
		p.cookiejar, err = cookiejar.New(nil)
		if err != nil {
			return err
		}
	}

	p.client = resty.New().
		SetTimeout(p.configuration.Endpoint.timeout).
		SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: p.configuration.TLS.InsecureSkipVerify,
			Certificates:       certificates,
		}).
		SetHeaders(p.configuration.Headers)

	if p.configuration.Session.UseCookieJar {
		p.client.SetCookieJar(p.cookiejar)
	}

	if p.configuration.TLS.RootCAFile != "" {
		p.client.SetRootCertificate(p.configuration.TLS.RootCAFile)
	}

	switch p.configuration.Authentication.Type {
	case BasicAuth:
		p.client.SetBasicAuth(p.configuration.Authentication.Basic.Username,
			p.configuration.Authentication.Basic.Password)
	case BearerAuth:
		p.client.SetAuthToken(p.configuration.Authentication.Bearer.Token)
	case APIKeyAuth:
		p.client.SetHeader(p.configuration.Authentication.APIKey.Header,
			p.configuration.Authentication.APIKey.Value)
	case DigestAuth:
		p.client.SetDigestAuth(p.configuration.Authentication.Digest.Username,
			p.configuration.Authentication.Digest.Password)
	default:

	}

	if p.configuration.Endpoint.Method == "" {
		p.configuration.Endpoint.Method = POST
	}

	return nil

}

func (p *Plugin) Produce(k []byte, v []byte, headers map[string]string) (*jrpc.ProduceResponse, error) {

	var err error

	// creating request
	req := p.client.R().
		SetBody(v)

	var resp *resty.Response

	switch p.configuration.Endpoint.Method {
	case POST:
		resp, err = req.Post(p.configuration.Endpoint.URL)
	case PUT:
		resp, err = req.Put(p.configuration.Endpoint.URL)
	default:
		resp, err = req.Post(p.configuration.Endpoint.URL)
	}

	if err != nil {
		return nil, err
	}
	defer resp.RawResponse.Body.Close()

	body, err := ioutil.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != p.configuration.ErrorHandling.ExpectStatusCode &&
		!p.configuration.ErrorHandling.IgnoreStatusCode {
		return nil, fmt.Errorf("Unexpected status code: %d", resp.StatusCode())
	}

	return &jrpc.ProduceResponse{
		Bytes:   uint64(len(v)),
		Message: string(body),
	}, nil

}

func (p *Plugin) Close(_ context.Context) error {
	return nil
}

func (p *Plugin) GetClient() *resty.Client {
	return p.client
}
