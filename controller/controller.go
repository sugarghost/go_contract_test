package controller

import (
	"go-contract/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	md *model.Model
}

func NewCTL(rep *model.Model) (*Controller, error) {
	r := &Controller{md: rep}
	return r, nil
}

func (p *Controller) GetOK(c *gin.Context) {
	c.JSON(200, gin.H{"msg": "ok"})
	return
}

func (p *Controller) SearchTokenSymbolByTokenNameController(c *gin.Context) {
	tokenName := c.Query("tokenName")
	symbol, err := p.md.SearchTokenSymbolByTokenNameModel(tokenName)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "symbol을 가져오지 못했습니다!",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"symbol": symbol})
}

func (p *Controller) SearchTokenBalanceByAddressController(c *gin.Context) {

	address := c.GetHeader("address")
	if address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "address 정보가 유효하지 않습니다",
		})
		return
	}

	balance, err := p.md.SearchTokenBalanceByAddressModel(address)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "balance를 가져오지 못했습니다!",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"balance": balance})
}

func (p *Controller) SendTokenByAddressController(c *gin.Context) {
	address := c.GetHeader("address")
	if address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "address 정보가 유효하지 않습니다",
		})
		return
	}
	err := p.md.SendTokenByAddressModel(address, "")

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "전송에 실패했습니다!",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"msg": "ok"})
}

func (p *Controller) SendTokenByAddressWithPrivateKeyController(c *gin.Context) {

	address := c.GetHeader("address")
	if address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "address 정보가 유효하지 않습니다",
		})
		return
	}
	privateKey := c.GetHeader("privateKey")
	if address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "privateKey 정보가 유효하지 않습니다",
		})
		return
	}
	err := p.md.SendTokenByAddressModel(address, privateKey)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "전송에 실패했습니다!",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"msg": "ok"})
}

func (p *Controller) SendWemixCoinByAddressController(c *gin.Context) {

	address := c.GetHeader("address")
	if address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "address 정보가 유효하지 않습니다",
		})
		return
	}
	err := p.md.SendWemixCoinByAddressModel(address, "")

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "전송에 실패했습니다!",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"msg": "ok"})
}

func (p *Controller) SendWemixCoinByAddressWithPrivateKeyController(c *gin.Context) {

	address := c.GetHeader("address")
	if address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "address 정보가 유효하지 않습니다",
		})
		return
	}
	privateKey := c.GetHeader("privateKey")
	if address == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "privateKey 정보가 유효하지 않습니다",
		})
		return
	}
	err := p.md.SendWemixCoinByAddressModel(address, privateKey)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "전송에 실패했습니다!",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{"msg": "ok"})
}
