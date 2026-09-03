package quickcmd

import "strings"

// DetectOSTab determines which OS runbook tab to open based on the detected host OS.
func DetectOSTab(distro string) OSTab {
	d := strings.ToLower(distro)
	switch {
	case strings.Contains(d, "ubuntu"), strings.Contains(d, "debian"), strings.Contains(d, "mint"):
		return OSTabUbuntu
	case strings.Contains(d, "rhel"), strings.Contains(d, "red hat"), strings.Contains(d, "centos"), strings.Contains(d, "rocky"), strings.Contains(d, "almalinux"), strings.Contains(d, "fedora"):
		return OSTabRHEL
	case strings.Contains(d, "alpine"):
		return OSTabAlpine
	default:
		return OSTabCommon
	}
}
