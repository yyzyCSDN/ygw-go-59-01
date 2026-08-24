package web

import _ "embed"

//go:embed monitor.html
var monitorHTML []byte

// MonitorHTML returns the embedded browser monitoring page.
func MonitorHTML() []byte {
	return monitorHTML
}
