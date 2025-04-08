package models

import "qlova.tech/sum"

type Screen struct {
	MainMenu,
	SectionSelection,
	ItemList,
	Loading,
	SearchBox,
	Download,
	DownloadArt sum.Int[Screen]
}
