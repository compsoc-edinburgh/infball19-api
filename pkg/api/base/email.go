package base

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	qrcode "github.com/skip2/go-qrcode"
	mailgun "gopkg.in/mailgun/mailgun-go.v1"
)

type EmailStruct struct {
	Name      string
	AuthToken string
	OrderID   string
}

var Email *template.Template = template.Must(template.ParseFiles("email.html"))

func SendTicketEmail(c *gin.Context, mailgun mailgun.Mailgun, name, to_address, orderID, authToken, qrCodeLocation string) (_ bool) {
	var tpl bytes.Buffer

	qrcode.WriteFile(authToken, qrcode.Medium, 512, qrCodeLocation+authToken+".png")

	if err := Email.Execute(&tpl, EmailStruct{
		Name:      name,
		OrderID:   orderID,
		AuthToken: authToken,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Ticket was bought but email generation was unsuccessful. Please email us for assistance.",
		})
		return
	}

	message := mailgun.NewMessage(
		"Infball <infball@comp-soc.com>",
		"Informatics ball 2019! [#"+orderID+"]",
		"",
		to_address,
	)
	message.SetHtml(tpl.String())

	_, _, err := mailgun.Send(message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Ticket was bought but an email was not sent. Please email us at infball@comp-soc.com for assistance.",
		})
		return
	}

	return true
}
