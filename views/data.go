package views

import (
	"log"

	"GoBlog/models"
)

const (
	AlertLvlError   = "red"
	AlertLvlWarning = "amber"
	AlertLvlInfo    = "blue"
	AlertLvlSuccess = "green"

	// AlertMessageGeneric is displayed when any
	// random error is encountered by our backend
	AlertMsgGeneric = "Something went wrong. Please try again or contact support "
)

type PublicError interface {
	error
	Public() string
}

// Data is the top level structure that views expect data to come in.
type Data struct {
	Alert   *Alert
	Account *models.Account
	Yield   interface{}
}

// Alert is used to render materialize css alert messages in templates.
type Alert struct {
	Level   string
	Message string
}

func (d *Data) AlertError(msg string) {
	d.Alert = &Alert{
		Level:   AlertLvlError,
		Message: msg,
	}
}

func (d *Data) SetAlert(err error) {
	var msg string
	if pErr, ok := err.(PublicError); ok {
		msg = pErr.Public()
	} else {
		log.Println(err)
		msg = AlertMsgGeneric
	}
	d.Alert = &Alert{
		Level:   AlertLvlError,
		Message: msg,
	}
}
