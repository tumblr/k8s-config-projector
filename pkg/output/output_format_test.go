package output

import (
	"fmt"
	"testing"
	"time"
)

func TestBuildFileOutputName(t *testing.T) {
	namespace := "testnamespace"
	name := "fakename"
	timestamp := time.Now().Unix()

	s := BuildFileOutputName(namespace, name, timestamp)

	if s != fmt.Sprintf("testnamespace--fakename--%d.yaml", timestamp) {
		t.Fatalf("File Output naming failed, expected output in the form of '%s--%s--%d', received '%s'", namespace, name, timestamp, s)
	}
}

func TestFormatName(t *testing.T) {
	underscoreName := "production-dc_twem.yaml"
	if s := FormatName(underscoreName); s != "production-dc-twem.yaml" {
		t.Fatalf("Failed testing underscoreName, expected production-dc-twem.yaml but received %s", s)
	}

	starName := "production-dc*twem.yaml"
	if s := FormatName(starName); s != "production-dc-twem.yaml" {
		t.Fatalf("Failed testing starName, expected production-dc-twem.yaml but received %s", s)
	}

	sameName := "production-dc-twem.yaml"
	if s := FormatName(sameName); s != "production-dc-twem.yaml" {
		t.Fatalf("Failed testing sameName, expected production-dc-twem.yaml but received %s", s)
	}
}
