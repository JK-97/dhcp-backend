package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/coreos/etcd/clientv3"
)

func getEndpoint() string {
	endpoint := os.Getenv("ETCD_ENDPOINT")
	if endpoint == "" {
		endpoint = "10.55.2.114:2379"
	}
	return endpoint
}

func TestDomainToKey(t *testing.T) {
	agent := EtcdDNSAgent{
		Prefix: "/skydns",
	}

	if agent.DomainToKey("skydns.local") != "/skydns/local/skydns/" {
		t.Error("Except:", "/skydns/local/skydns/", "Actually:", agent.DomainToKey("skydns.local"))
	}

	if agent.DomainToKey("skydns.local.") != "/skydns/local/skydns/" {
		t.Error("Except:", "/skydns/local/skydns/", "Actually:", agent.DomainToKey("skydns.local."))
	}
}

func TestKeyToDomain(t *testing.T) {
	agent := EtcdDNSAgent{
		Prefix: "/skydns",
	}

	key := "/skydns/local/skydns/"
	if agent.KeyToDomain(key) != "skydns.local" {
		t.Error("Except:", "skydns.local", "Actually:", agent.KeyToDomain(key))
	}

}

func TestAddTxtRecord(t *testing.T) {
	kvc, _ := clientv3.New(clientv3.Config{
		Endpoints: []string{
			getEndpoint(),
		}})

	agent := EtcdDNSAgent{
		cli:    kvc,
		Prefix: "/skydns",
	}

	r := DNSRecord{
		Name: "jx.skydns.local",
		Type: "SRV",
		Host: "127.0.0.1",
		Port: 8080,
	}

	err := agent.AddRecord(r)

	if err != nil {
		t.Error(err)
	}
}
func TestAddTxtRecordHTTP(t *testing.T) {
	kvc, err := clientv3.New(clientv3.Config{
		Endpoints: []string{
			getEndpoint(),
		}})

	if err != nil {
		panic(err)
	}

	agent := EtcdDNSAgent{
		cli:    kvc,
		Prefix: "/skydns",
	}

	recorder := httptest.NewRecorder()
	rec := DNSRecord{
		Name:     "jx.skydns.local",
		Type:     "TXT",
		Host:     "127.0.0.1",
		Priority: 10,
		TTL:      10,
		Weight:   10,
		Port:     8080,
	}
	b, err := json.Marshal(rec)
	if err != nil {
		t.Error(err)
	}

	t.Log(string(b))
	body := bytes.NewReader(b)
	r, err := http.NewRequest(http.MethodPost, "", body)
	if err != nil {
		t.Error(err)
	}
	backend := DNSBackend{Agent: &agent}
	backend.AddRecord(recorder, r)

	if recorder.Code != http.StatusOK {
		t.Error(recorder.Code)
	}
	// backend.AddRecord()
}
