package charge

import (
	"fmt"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/go-redis/redis"

	"github.com/compsoc-edinburgh/infball19-api/pkg/api/base"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	stripe "github.com/stripe/stripe-go"
)

func (i *Impl) MakeCharge(c *gin.Context) {
	var result struct {
		Token     string
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
		base.BadRequest(c, "Invalid staff code provided.")
		return
	}

	if result.Token == "" {
		base.BadRequest(c, "Card information is missing.")
		return
	}

	// if !result.Over18 {
	// 	base.BadRequest(c, "You must be atleast 18 years of age to attend.")
	// 	return
	// }

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

	if !base.CheckUUN(c, result.UUN) && result.StaffCode != "" && result.StaffCode != i.Config.StaffCode {
		base.BadRequest(c, "Invalid uun provided")
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
	sku := new(stripe.SKU)
	skuString := ""
	if !result.NoAlcohol {
		fmt.Println("lets get wasted")
		sku, err = i.Stripe.Skus.Get(i.Config.Stripe.SKU, nil)
		skuString = i.Config.Stripe.SKU
	} else {
		fmt.Println("bismillah")
		sku, err = i.Stripe.Skus.Get(i.Config.Stripe.NonAlcoholicSKU, nil)
		skuString = i.Config.Stripe.NonAlcoholicSKU
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

	authToken := uuid.New().String()

	order, err := i.Stripe.Orders.New(&stripe.OrderParams{
		Currency: stripe.String(string(stripe.CurrencyGBP)),
		Items: []*stripe.OrderItemParams{
			&stripe.OrderItemParams{
				Type:   stripe.String(string(stripe.OrderItemTypeSKU)),
				Parent: stripe.String(skuString),
			},
		},
		Params: stripe.Params{
			Metadata: map[string]string{
				"uun":             result.UUN,
				"purchaser_email": result.Email,
				"purchaser_name":  result.FullName,
				"owner_email":     result.Email,
				"owner_name":      result.FullName,
				// "over18":          strconv.FormatBool(result.Over18),
				"meal_types":       strings.Join(result.MealTypes[:], ","),
				"special_requests": result.SpecialReqs,
				"auth_token":       authToken,
			},
		},
		Email: stripe.String(result.Email),
	})

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

	// Charge the user's card:
	params := &stripe.OrderPayParams{}
	params.SetSource(result.Token)

	// Actually pay the user
	o, err := i.Stripe.Orders.Pay(order.ID, params)
	if err != nil {
		msg := err.Error()
		if stripeErr, ok := err.(*stripe.Error); ok {
			msg = stripeErr.Msg
		}

		i.Stripe.Orders.Update(order.ID, &stripe.OrderUpdateParams{
			Status: stripe.String(string(stripe.OrderStatusCanceled)),
		})

		base.BadRequest(c, msg)
		return
	}

	go i.Stripe.Charges.Update(o.Charge.ID, &stripe.ChargeParams{
		Description: stripe.String("Informatics ball 2019 ticket"),
	})

	client := redis.NewClient(&redis.Options{
		Addr:     i.Config.Redis.Address,
		Password: i.Config.Redis.Password,
		DB:       i.Config.Redis.DB,
	})
	err = client.Set(authToken, result.FullName, 0).Err()
	if err != nil {
		//base.BadRequest(c, "An internal database error occured, although your card has been charged")
		//return
		fmt.Println(err)
	}

	if !base.SendTicketEmail(c, i.Mailgun, result.FullName, toAddress, o.ID, authToken) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   o.ID,
	})
}
