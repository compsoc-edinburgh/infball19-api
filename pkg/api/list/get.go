package list

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	stripe "github.com/stripe/stripe-go"
)

func (i *Impl) Get(c *gin.Context) {

	params := &stripe.OrderListParams{}
	params.AddExpand("data.charge.balance_transaction")
	params.Filters.AddFilter("status", "", "paid")
	sku := ""

	// day before orders went out, utc timestamp
	params.Filters.AddFilter("created", "gt", "1518714907")

	file, _ := os.Create("attendees.csv")
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{
		"order_id",
		"owner_name", "owner_email",
		"uun",
		"meal_type",
		"special_requests",
		"purchaser_name", "purchaser_email",
		// "over18",
		"auth_token",
		"charge_net",
		"charge_fees",
		"no_alcohol",
	})

	orders := i.Stripe.Orders.List(params)
	for orders.Next() {
		o := orders.Order()

		hasSKU := false
		for _, item := range o.Items {
			// todo: check if item.Parent always exists
			fmt.Println(item.Parent)
			if item.Parent != nil {
				if item.Parent.ID == i.Config.Stripe.SKU || item.Parent.ID == i.Config.Stripe.NonAlcoholicSKU {
					sku = item.Parent.ID
					hasSKU = true
					continue
				}
			}
		}

		if !hasSKU {
			continue
		}

		writer.Write([]string{
			o.ID,
			o.Metadata["owner_name"], o.Metadata["owner_email"],
			o.Metadata["uun"],
			o.Metadata["meal_types"],
			o.Metadata["special_requests"],
			o.Metadata["purchaser_name"], o.Metadata["purchaser_email"],
			//o.Metadata["over18"],
			o.Metadata["auth_token"],
			strconv.FormatInt(o.Charge.BalanceTransaction.Net, 10),
			strconv.FormatInt(o.Charge.BalanceTransaction.Fee, 10),
			strconv.FormatBool(sku == i.Config.Stripe.NonAlcoholicSKU),
		})
	}

	// writer.Flush()

	c.String(http.StatusOK, "ok")
}
