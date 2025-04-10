package common

import (
	"os"
	"path/filepath"
)

func ConfigureEnvironment() {
	cwd, err := os.Getwd()
	if err != nil {
		LogStandardFatal("Failed to get current working directory", err)
	}

	if err := os.Setenv("DEVICE", "brick"); err != nil {
		LogStandardFatal("Failed to set DEVICE", err)
	}
	if err := os.Setenv("PLATFORM", "tg5040"); err != nil {
		LogStandardFatal("Failed to set PLATFORM", err)
	}
	if err := os.Setenv("PATH", filepath.Join(cwd, "bin/tg5040")); err != nil {
		LogStandardFatal("Failed to set PATH", err)
	}
	if err := os.Setenv("LD_LIBRARY_PATH", "/mnt/SDCARD/.system/tg5040/lib:/usr/trimui/lib"); err != nil {
		LogStandardFatal("Failed to set LD_LIBRARY_PATH", err)
	}

	// So users don't have to install TrimUI_EX
	if err := os.Setenv("SSL_CERT_DIR", filepath.Join(cwd, "certs")); err != nil {
		LogStandardFatal("Failed to set SSL_CERT_DIR", err)
	}
}
