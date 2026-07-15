package app

type Platform interface {
	MachineID() string
	IsCertInstalled(certPath string) bool
	InstallCert(certPath string) bool
	RemoveCert()
	SetPAC(pacURL string)
	UnsetPAC()
	OpenBrowser(url string)
	ProcessRunning(pid int) bool
}
