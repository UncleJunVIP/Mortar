module github.com/UncleJunVIP/Mortar.pak

go 1.24.1

replace github.com/UncleJunVIP/nextui-pak-shared-functions => ../nextui-pak-shared-functions // TODO remove this before committing!

require (
	github.com/UncleJunVIP/nextui-pak-shared-functions v0.0.0-00010101000000-000000000000
	github.com/activcoding/HTML-Table-to-JSON v0.0.4
	github.com/hirochachacha/go-smb2 v1.1.0
	go.uber.org/atomic v1.11.0
	go.uber.org/zap v1.27.0
	golang.org/x/text v0.23.0
	gopkg.in/yaml.v3 v3.0.1
	qlova.tech v0.1.1
)

require (
	github.com/PuerkitoBio/goquery v1.9.2 // indirect
	github.com/andybalholm/cascadia v1.3.2 // indirect
	github.com/geoffgarside/ber v1.1.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/net v0.37.0 // indirect
)
