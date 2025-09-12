package ui

import (
	"mortar/models"

	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"qlova.tech/sum"
)

type Search struct {
	Platform    models.Platform
	InitialText string
}

func InitSearch(platform models.Platform, initialText string) Search {
	return Search{
		Platform:    platform,
		InitialText: initialText,
	}
}

func (s Search) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.SearchBox
}

func (s Search) Draw() (value interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	res, err := gabagool.Keyboard(s.InitialText)
	if err != nil {
		logger.Error("Error with blocking keyboard", "error", err)
		return nil, -1, err
	}

	if res.IsSome() {
		return res.Unwrap(), 0, nil
	}

	return nil, -1, nil
}
