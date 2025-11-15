package web

import (
	"fmt"
	"mortar/utils"

	"github.com/UncleJunVIP/gabagool/pkg/gabagool"
)

func QRScreen(helpText string) {

	ip, err := utils.GetLocalIP()
	if err != nil {
		return
	}

	qrURL := fmt.Sprintf("https://mortar.unclejun.vip?api=%s", ip)

	tmpQR, err := utils.CreateTempQRCode(qrURL, 128)
	if err != nil {
		return
	}

	start()

	message := fmt.Sprintf(qrURL)

	gabagool.ConfirmationMessage(message, []gabagool.FooterHelpItem{
		{ButtonName: "A", HelpText: helpText},
	}, gabagool.MessageOptions{
		ImagePath: tmpQR,
	})
	stop()
}
