package models

import "qlova.tech/sum"

type ScreenName struct {
	MainMenu,
	Settings,
	PlatformSelection,
	GameList,
	GlobalGamesList,
	SearchBox,
	Download,
	DownloadArt sum.Int[ScreenName]
}

var ScreenNames = sum.Int[ScreenName]{}.Sum()

type Screen interface {
	Name() sum.Int[ScreenName]
	Draw() (value interface{}, exitCode int, e error)
}
