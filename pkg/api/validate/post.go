package validate

import (
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/compsoc-edinburgh/infball19-api/pkg/api/base"
	"github.com/gin-gonic/gin"
	stripe "github.com/stripe/stripe-go"
)

func (i *Impl) Post(c *gin.Context) {
	var result struct {
		StaffCode string // special code

		FullName string
		UUN      string
		Email    string
		// Over18      bool
		MealTypes []string
		NoAlcohol bool

		SpecialReqs string
	}
	fmt.Println(c.ContentType())

	if err := c.BindJSON(&result); err != nil {
		base.BadRequest(c, err.Error())
		return
	}

	currentTime := time.Now()
	closeTime, _ := time.Parse(time.RFC3339, "2019-03-21T00:00:00+00:00")

	if currentTime.Unix() > closeTime.Unix() {
		base.BadRequest(c, "Ticket sales have now closed.")
		return
	}

	if result.StaffCode != "" && result.StaffCode != i.Config.StaffCode {
		base.BadRequest(c, "Invalid invite code provided.")
		return
	}

	if result.FullName == "" {
		base.BadRequest(c, "Full name missing.")
		return
	}

	toAddress := result.FullName + "<" + result.Email + ">"
	_, err := mail.ParseAddress(toAddress)
	if err != nil {
		base.BadRequest(c, "Invalid email format provided. Please email infball@comp-soc.com if this is a mistake.")
		return
	}

	if !base.CheckUUN(c, result.UUN) && result.StaffCode != i.Config.StaffCode {
		base.BadRequest(c, "You are not authorised to buy a ticket as your uun or invite code is invalid, if this is a mistake please contact us at infball@comp-soc.com")
		return
	}

	if !base.IsMealValid(result.MealTypes) {
		base.BadRequest(c, "Invalid food selection.")
		return
	}
	if len(result.SpecialReqs) > 500 {
		base.BadRequest(c, "Sorry, your request is limited to 500 characters. Please email infball@comp-soc.com for assistance.")
		return
	}

	sku, err := i.Stripe.Skus.Get(i.Config.Stripe.SKU, nil)
	if result.NoAlcohol == true {
		sku, err = i.Stripe.Skus.Get(i.Config.Stripe.NonAlcoholicSKU, nil)
	}
	if err != nil {
		msg := err.Error()
		if stripeErr, ok := err.(*stripe.Error); ok {
			msg = stripeErr.Msg
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": msg,
		})
		return
	}

	fmt.Printf("%+v", sku.Inventory)
	if sku.Inventory.Quantity == 0 {
		c.JSON(http.StatusGone, gin.H{
			"status":  "error",
			"message": "Sorry! We have run out of those tickets... for now.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
	return
}
