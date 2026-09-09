package profile

import (
	"testing"
)

func TestProfileDecryption(t *testing.T) {
	koProf, links := GetProfile("ko")
	if koProf.Name != "김경곤 (Kyunggon Kim)" {
		t.Fatalf("unexpected Korean profile name: %s", koProf.Name)
	}
	if links.Web != "https://retry.day" {
		t.Fatalf("unexpected web link: %s", links.Web)
	}

	enProf, _ := GetProfile("en")
	if enProf.Name != "Kyunggon Kim" {
		t.Fatalf("unexpected English profile name: %s", enProf.Name)
	}

	deProf, _ := GetProfile("de")
	if deProf.Name != "Kyunggon Kim" {
		t.Fatalf("unexpected German profile name: %s", deProf.Name)
	}
}
