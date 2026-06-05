package openstack

import (
	"context"
	"testing"

	"github.com/dihedron/devws/test"
)

func TestList(t *testing.T) {
	test.Setup(t)
	client, err := NewClient(t.Context())
	if err != nil {
		t.Logf("error creating client: %v", err)
		t.FailNow()
	}
	servers, err := client.ComputeV2.List(context.TODO())
	if err != nil {
		t.Logf("error listing servers: %v", err)
		t.FailNow()
	}
	for _, server := range servers {
		t.Logf("server (id: %s, name: %s)", server.ID, server.Name)
	}
}
