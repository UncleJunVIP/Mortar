package ui

import (
	"mortar/models"
	"qlova.tech/sum"
)

var Screens = sum.Int[models.ScreenName]{}.Sum()
