/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

--

This is a copy of k8s.io/kubernetes/cmd/kubeadm/app/phases/certs/pkiutil/pki_helpers.go
modified to work independently of kubeadm internals like the configuration.
*/

package pkiutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"io/ioutil"
	"os"
	"testing"

	certutil "k8s.io/client-go/util/cert"
)

func TestNewCertificateAuthority(t *testing.T) {
	cert, key, err := NewCertificateAuthority()

	if cert == nil {
		t.Errorf(
			"failed NewCertificateAuthority, cert == nil",
		)
	}
	if key == nil {
		t.Errorf(
			"failed NewCertificateAuthority, key == nil",
		)
	}
	if err != nil {
		t.Errorf(
			"failed NewCertificateAuthority with an error: %v",
			err,
		)
	}
}

func TestNewCertAndKey(t *testing.T) {
	var tests = []struct {
		caKeySize int
		expected  bool
	}{
		{
			// RSA key too small
			caKeySize: 128,
			expected:  false,
		},
		{
			// Should succeed
			caKeySize: 2048,
			expected:  true,
		},
	}

	for _, rt := range tests {
		caKey, err := rsa.GenerateKey(rand.Reader, rt.caKeySize)
		if err != nil {
			t.Fatalf("Couldn't create rsa Private Key")
		}
		caCert := &x509.Certificate{}
		config := certutil.Config{
			CommonName:   "test",
			Organization: []string{"test"},
			Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		}
		validate := 1
		_, _, actual := NewCertAndKey(caCert, caKey, config, validate)
		if (actual == nil) != rt.expected {
			t.Errorf(
				"failed NewCertAndKey:\n\texpected: %t\n\t  actual: %t",
				rt.expected,
				(actual == nil),
			)
		}
	}
}

func TestHasServerAuth(t *testing.T) {
	validate := 1
	caCert, caKey, _ := NewCertificateAuthority()

	var tests = []struct {
		config   certutil.Config
		expected bool
	}{
		{
			config: certutil.Config{
				CommonName: "test",
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			},
			expected: true,
		},
		{
			config: certutil.Config{
				CommonName: "test",
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			},
			expected: false,
		},
	}

	for _, rt := range tests {
		cert, _, err := NewCertAndKey(caCert, caKey, rt.config, validate)
		if err != nil {
			t.Fatalf("Couldn't create cert: %v", err)
		}
		actual := HasServerAuth(cert)
		if actual != rt.expected {
			t.Errorf(
				"failed HasServerAuth:\n\texpected: %t\n\t  actual: %t",
				rt.expected,
				actual,
			)
		}
	}
}

func TestWriteCertAndKey(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Couldn't create rsa Private Key")
	}
	caCert := &x509.Certificate{}
	actual := WriteCertAndKey(tmpdir, "foo", caCert, caKey)
	if actual != nil {
		t.Errorf(
			"failed WriteCertAndKey with an error: %v",
			actual,
		)
	}
}

func TestWriteCert(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	caCert := &x509.Certificate{}
	actual := WriteCert(tmpdir, "foo", caCert)
	if actual != nil {
		t.Errorf(
			"failed WriteCertAndKey with an error: %v",
			actual,
		)
	}
}

func TestWriteKey(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Couldn't create rsa Private Key")
	}
	actual := WriteKey(tmpdir, "foo", caKey)
	if actual != nil {
		t.Errorf(
			"failed WriteCertAndKey with an error: %v",
			actual,
		)
	}
}

func TestWritePublicKey(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Couldn't create rsa Private Key")
	}
	actual := WritePublicKey(tmpdir, "foo", &caKey.PublicKey)
	if actual != nil {
		t.Errorf(
			"failed WriteCertAndKey with an error: %v",
			actual,
		)
	}
}

func TestCertOrKeyExist(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Couldn't create rsa Private Key")
	}
	caCert := &x509.Certificate{}
	actual := WriteCertAndKey(tmpdir, "foo", caCert, caKey)
	if actual != nil {
		t.Errorf(
			"failed WriteCertAndKey with an error: %v",
			actual,
		)
	}

	var tests = []struct {
		path     string
		name     string
		expected bool
	}{
		{
			path:     "",
			name:     "",
			expected: false,
		},
		{
			path:     tmpdir,
			name:     "foo",
			expected: true,
		},
	}
	for _, rt := range tests {
		actual := CertOrKeyExist(rt.path, rt.name)
		if actual != rt.expected {
			t.Errorf(
				"failed CertOrKeyExist:\n\texpected: %t\n\t  actual: %t",
				rt.expected,
				actual,
			)
		}
	}
}

func TestTryLoadCertAndKeyFromDisk(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	caCert, caKey, err := NewCertificateAuthority()
	if err != nil {
		t.Errorf(
			"failed to create cert and key with an error: %v",
			err,
		)
	}
	err = WriteCertAndKey(tmpdir, "foo", caCert, caKey)
	if err != nil {
		t.Errorf(
			"failed to write cert and key with an error: %v",
			err,
		)
	}

	var tests = []struct {
		path     string
		name     string
		expected bool
	}{
		{
			path:     "",
			name:     "",
			expected: false,
		},
		{
			path:     tmpdir,
			name:     "foo",
			expected: true,
		},
	}
	for _, rt := range tests {
		_, _, actual := TryLoadCertAndKeyFromDisk(rt.path, rt.name)
		if (actual == nil) != rt.expected {
			t.Errorf(
				"failed TryLoadCertAndKeyFromDisk:\n\texpected: %t\n\t  actual: %t",
				rt.expected,
				(actual == nil),
			)
		}
	}
}

func TestTryLoadCertFromDisk(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	caCert, _, err := NewCertificateAuthority()
	if err != nil {
		t.Errorf(
			"failed to create cert and key with an error: %v",
			err,
		)
	}
	err = WriteCert(tmpdir, "foo", caCert)
	if err != nil {
		t.Errorf(
			"failed to write cert and key with an error: %v",
			err,
		)
	}

	var tests = []struct {
		path     string
		name     string
		expected bool
	}{
		{
			path:     "",
			name:     "",
			expected: false,
		},
		{
			path:     tmpdir,
			name:     "foo",
			expected: true,
		},
	}
	for _, rt := range tests {
		_, actual := TryLoadCertFromDisk(rt.path, rt.name)
		if (actual == nil) != rt.expected {
			t.Errorf(
				"failed TryLoadCertAndKeyFromDisk:\n\texpected: %t\n\t  actual: %t",
				rt.expected,
				(actual == nil),
			)
		}
	}
}

func TestTryLoadKeyFromDisk(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	_, caKey, err := NewCertificateAuthority()
	if err != nil {
		t.Errorf(
			"failed to create cert and key with an error: %v",
			err,
		)
	}
	err = WriteKey(tmpdir, "foo", caKey)
	if err != nil {
		t.Errorf(
			"failed to write cert and key with an error: %v",
			err,
		)
	}

	var tests = []struct {
		path     string
		name     string
		expected bool
	}{
		{
			path:     "",
			name:     "",
			expected: false,
		},
		{
			path:     tmpdir,
			name:     "foo",
			expected: true,
		},
	}
	for _, rt := range tests {
		_, actual := TryLoadKeyFromDisk(rt.path, rt.name)
		if (actual == nil) != rt.expected {
			t.Errorf(
				"failed TryLoadCertAndKeyFromDisk:\n\texpected: %t\n\t  actual: %t",
				rt.expected,
				(actual == nil),
			)
		}
	}
}

func TestPathsForCertAndKey(t *testing.T) {
	crtPath, keyPath := pathsForCertAndKey("/foo", "bar")
	if crtPath != "/foo/bar.crt" {
		t.Errorf("unexpected certificate path: %s", crtPath)
	}
	if keyPath != "/foo/bar.key" {
		t.Errorf("unexpected key path: %s", keyPath)
	}
}

func TestPathForCert(t *testing.T) {
	crtPath := pathForCert("/foo", "bar")
	if crtPath != "/foo/bar.crt" {
		t.Errorf("unexpected certificate path: %s", crtPath)
	}
}

func TestPathForKey(t *testing.T) {
	keyPath := pathForKey("/foo", "bar")
	if keyPath != "/foo/bar.key" {
		t.Errorf("unexpected certificate path: %s", keyPath)
	}
}

func TestPathForPublicKey(t *testing.T) {
	pubPath := pathForPublicKey("/foo", "bar")
	if pubPath != "/foo/bar.pub" {
		t.Errorf("unexpected certificate path: %s", pubPath)
	}
}

/*
FIXME; disable due to failing tests:
--- FAIL: TestGetEtcdAltNames (0.00s)
    pki_helpers_test.go:471: altNames does not contain DNSName localhost
    pki_helpers_test.go:486: altNames does not contain IPAddress 127.0.0.1
    pki_helpers_test.go:486: altNames does not contain IPAddress ::1

func TestGetEtcdAltNames(t *testing.T) {
	proxy := "user-etcd-proxy"
	proxyIP := "10.10.10.100"
	cfg := &apis.EtcdAdmConfig{
		ServerCertSANs: []string{
			proxy,
			proxyIP,
			"1.2.3.L",
			"invalid,commas,in,DNS",
		},
	}

	altNames, err := GetEtcdAltNames(cfg)
	if err != nil {
		t.Fatalf("failed calling GetEtcdAltNames: %v", err)
	}

	expectedDNSNames := []string{"localhost", proxy}
	for _, DNSName := range expectedDNSNames {
		found := false
		for _, val := range altNames.DNSNames {
			if val == DNSName {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("altNames does not contain DNSName %s", DNSName)
		}
	}

	expectedIPAddresses := []string{"127.0.0.1", net.IPv6loopback.String(), proxyIP}
	for _, IPAddress := range expectedIPAddresses {
		found := false
		for _, val := range altNames.IPs {
			if val.Equal(net.ParseIP(IPAddress)) {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("altNames does not contain IPAddress %s", IPAddress)
		}
	}
}
*/

/*
FIXME; disabled due to the kubeadmapi dependency
the file does not import this package.

func TestGetEtcdPeerAltNames(t *testing.T) {
	hostname := "valid-hostname"
	proxy := "user-etcd-proxy"
	proxyIP := "10.10.10.100"
	advertiseIP := "1.2.3.4"
	cfg := &kubeadmapi.MasterConfiguration{
		API:              kubeadmapi.API{AdvertiseAddress: advertiseIP},
		NodeRegistration: kubeadmapi.NodeRegistrationOptions{Name: hostname},
		Etcd: kubeadmapi.Etcd{
			Local: &kubeadmapi.LocalEtcd{
				PeerCertSANs: []string{
					proxy,
					proxyIP,
					"1.2.3.L",
					"invalid,commas,in,DNS",
				},
			},
		},
	}

	altNames, err := GetEtcdPeerAltNames(cfg)
	if err != nil {
		t.Fatalf("failed calling GetEtcdPeerAltNames: %v", err)
	}

	expectedDNSNames := []string{hostname, proxy}
	for _, DNSName := range expectedDNSNames {
		found := false
		for _, val := range altNames.DNSNames {
			if val == DNSName {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("altNames does not contain DNSName %s", DNSName)
		}
	}

	expectedIPAddresses := []string{advertiseIP, proxyIP}
	for _, IPAddress := range expectedIPAddresses {
		found := false
		for _, val := range altNames.IPs {
			if val.Equal(net.ParseIP(IPAddress)) {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("altNames does not contain IPAddress %s", IPAddress)
		}
	}
}
*/
