package quickcmd

import "testing"

func TestDetectOSTab(t *testing.T) {
	tests := []struct {
		distro   string
		expected OSTab
	}{
		{"Ubuntu 24.04.4 LTS", OSTabUbuntu},
		{"Debian GNU/Linux 12", OSTabUbuntu},
		{"Linux Mint 21", OSTabUbuntu},
		{"Rocky Linux 9.3", OSTabRHEL},
		{"CentOS Stream 9", OSTabRHEL},
		{"Red Hat Enterprise Linux 8.8", OSTabRHEL},
		{"Alpine Linux v3.19", OSTabAlpine},
		{"Arch Linux", OSTabCommon},
		{"", OSTabCommon},
	}

	for _, tt := range tests {
		got := DetectOSTab(tt.distro)
		if got != tt.expected {
			t.Errorf("DetectOSTab(%q) = %v, expected %v", tt.distro, got, tt.expected)
		}
	}
}

func TestCatalogIntegrity(t *testing.T) {
	for tab, items := range Catalog {
		if len(items) == 0 {
			t.Errorf("Catalog tab %v has no commands", tab)
		}
		for _, item := range items {
			if item.ID == "" || item.Command == "" || item.TitleKey == "" || item.DescKey == "" {
				t.Errorf("Catalog item has missing fields: %+v", item)
			}
		}
	}
}
