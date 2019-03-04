package qr

import (
	"github.com/gin-gonic/gin"

	qrcode "github.com/skip2/go-qrcode"
)

func (i *Impl) Get(c *gin.Context) {

	authToken := c.Param("auth_token")

	qr, _ := qrcode.Encode(authToken, qrcode.Medium, 512)
	//sEnc := b64.StdEncoding.EncodeToString(qr)

	c.Data(200, "image/png", qr)
	return

}
