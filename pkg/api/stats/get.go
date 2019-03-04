package stats

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (i *Impl) Get(c *gin.Context) {
	sku, err := i.Stripe.Skus.Get(i.Config.Stripe.SKU, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Unknown error.",
		})
	}

	c.JSON(http.StatusOK, gin.H{"quantity": sku.Inventory.Quantity})
}

func (i *Impl) GetNonAlcoholic(c *gin.Context) {
	sku, err := i.Stripe.Skus.Get(i.Config.Stripe.NonAlcoholicSKU, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Unknown error.",
		})
	}

	c.JSON(http.StatusOK, gin.H{"quantity": sku.Inventory.Quantity})
}
