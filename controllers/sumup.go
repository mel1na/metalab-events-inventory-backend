package controllers

import (
	"context"
	"fmt"
	"io"
	"metalab/events-inventory-tracker/models"
	sumup_models "metalab/events-inventory-tracker/models/sumup"
	"metalab/events-inventory-tracker/sumup_integration"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sumup/sumup-go"
)

func CreateReader(c *gin.Context) {
	var input sumup.CreateReaderBody
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if string(input.PairingCode) == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "missing pairing code"})
		return
	}

	reader, err := sumup_integration.SumupClient.Readers.Create(context.Background(), *sumup_integration.SumupAccount.MerchantProfile.MerchantCode, sumup.CreateReaderBody{Name: input.Name, PairingCode: sumup.ReaderPairingCode(input.PairingCode)})
	if err != nil {
		fmt.Printf("error while creating reader: %s\n", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//dbReader := sumup_models.Reader{ReaderID: string(reader.Id), ReaderName: string(reader.Name), ReaderDevice: sumup_models.ReaderDevice{Identifier: reader.Device.Identifier, Model: string(reader.Device.Model)}, ReaderStatus: string(reader.Status), ReaderCreatedAt: reader.CreatedAt, ReaderUpdatedAt: reader.UpdatedAt}
	db_reader := sumup_models.Reader{ReaderId: sumup_models.ReaderId(reader.Id), Name: sumup_models.ReaderName(reader.Name), Status: sumup_models.ReaderStatus(reader.Status), Device: sumup_models.ReaderDevice{Identifier: reader.Device.Identifier, Model: sumup_models.ReaderDeviceModel(reader.Device.Model)}, CreatedAt: reader.CreatedAt, UpdatedAt: reader.UpdatedAt}
	models.DB.Create(&db_reader)

	c.JSON(http.StatusOK, gin.H{"data": db_reader})
}

func CreateReaderCheckout(c *gin.Context) {
	var input sumup.CreateReaderCheckout
	var returnUrl string = os.Getenv("SUMUP_RETURN_URL")
	if input_err := c.ShouldBindJSON(&input); input_err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": input_err.Error()})
		return
	}

	db_reader, find_err := FindReaderByName("Bar")
	if find_err != nil {
		fmt.Printf("error finding reader by name: %s\n", find_err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": find_err.Error()})
		return
	}

	response, checkout_err := sumup_integration.SumupClient.Readers.CreateCheckout(context.Background(), *sumup_integration.SumupAccount.MerchantProfile.MerchantCode, string(db_reader.ReaderId), sumup.CreateReaderCheckoutBody{ReturnUrl: &returnUrl, TotalAmount: sumup.CreateReaderCheckoutAmount{Currency: "EUR", MinorUnit: 2, Value: input.TotalAmount.Value}})
	if checkout_err != nil {
		fmt.Printf("error while creating reader: %s\n", checkout_err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": checkout_err.Error()})
		return
	}

	//db_checkout := sumup.CreateReaderCheckout201Response{Data: &sumup.CreateReaderCheckout201ResponseData{ClientTransactionId: response.Data.ClientTransactionId}}
	//models.DB.Create(&db_checkout)

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func FindReaders(c *gin.Context) {
	var readers []sumup.Reader
	err := models.DB.Find(&readers).Error

	if err != nil {
		fmt.Printf("error finding readers: %s\n", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": readers})
}

func FindApiReaders(c *gin.Context) {
	response, err := sumup_integration.SumupClient.Readers.List(context.Background(), *sumup_integration.SumupAccount.MerchantProfile.MerchantCode)
	if err != nil {
		fmt.Printf("error finding reader by name: %s\n", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func FindReader(c *gin.Context) {
	var reader sumup_models.Reader

	if err := models.DB.Where("reader_id = ?", c.Param("id")).First(&reader).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": reader})
}

func FindReaderByName(name string) (*sumup_models.Reader, error) {
	var reader sumup_models.Reader

	if err := models.DB.Where("name = ?", name).First(&reader).Error; err != nil {
		return nil, err
	}

	return &reader, nil
}

func DeleteReaderById(id string) error {
	var reader sumup_models.Reader

	if err := models.DB.Where("reader_id = ?", id).Delete(&reader).Error; err != nil {
		return err
	}
	return nil
}

func DeleteReaderByName(name string) error {
	var reader sumup_models.Reader

	if err := models.DB.Where("name = ?", name).Delete(&reader).Error; err != nil {
		return err
	}
	return nil
}

func FindReaderIdByName(name string) *sumup_models.ReaderId {
	var reader sumup_models.Reader

	if err := models.DB.Where("name = ?", name).First(&reader).Error; err != nil {
		fmt.Printf("error finding reader id by name: %s\n", err.Error())
		return nil
	}

	return &reader.ReaderId
}

type TerminateReaderInput struct {
	ReaderId   string `json:"id"`
	ReaderName string `json:"name"`
}

// TODO: handle failed transactions via callback from sumup api
func TerminateReaderCheckout(c *gin.Context) {
	var input TerminateReaderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.ReaderId == "" && input.ReaderName == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "reader id/name missing"})
		return
	} else if input.ReaderId == "" && input.ReaderName != "" { //name defined, id undefined
		var db_reader *sumup_models.Reader
		var find_err error
		db_reader, find_err = FindReaderByName(input.ReaderId)
		if find_err != nil {
			fmt.Printf("error finding reader by name: %s\n", find_err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": find_err.Error()})
			return
		}

		terminate_err := sumup_integration.SumupClient.Readers.TerminateCheckout(context.Background(), *sumup_integration.SumupAccount.MerchantProfile.MerchantCode, string(db_reader.ReaderId)) //uses reader id from db, retrieved from name
		if terminate_err != nil {
			fmt.Printf("error while terminating checkout by name: %s\n", terminate_err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": terminate_err.Error()})
			return
		}
	} else if input.ReaderId != "" && input.ReaderName == "" { //name undefined, id defined
		terminate_err := sumup_integration.SumupClient.Readers.TerminateCheckout(context.Background(), *sumup_integration.SumupAccount.MerchantProfile.MerchantCode, input.ReaderId) // uses reader id from input
		if terminate_err != nil {
			fmt.Printf("error while terminating checkout by id: %s\n", terminate_err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": terminate_err.Error()})
			return
		}
	} else {
		fmt.Printf("unknown error while terminating checkout\n")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unknown error while terminating checkout"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": "success"})
}

type UnlinkReaderInput struct {
	ReaderId   string `json:"id"`
	ReaderName string `json:"name"`
}

func UnlinkReader(c *gin.Context) {
	var input UnlinkReaderInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.ReaderId == "" && input.ReaderName == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "reader id/name missing"})
		return
	} else if input.ReaderId == "" && input.ReaderName != "" { //name defined
		var db_reader *sumup_models.Reader
		var find_err error
		db_reader, find_err = FindReaderByName(input.ReaderName)
		if find_err != nil {
			fmt.Printf("error finding reader by name: %s\n", find_err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": find_err.Error()})
			return
		}

		unlink_err := sumup_integration.SumupClient.Readers.DeleteReader(context.Background(), *sumup_integration.SumupAccount.MerchantProfile.MerchantCode, sumup.ReaderId(db_reader.ReaderId))
		if unlink_err != nil {
			fmt.Printf("error while unlinking reader by name: %s\n", unlink_err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": unlink_err.Error()})
			return
		}
		if delete_err := DeleteReaderByName(input.ReaderName); delete_err != nil {
			fmt.Printf("error while deleting reader by name: %s\n", delete_err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": delete_err.Error()})
		}
	} else if input.ReaderId != "" && input.ReaderName == "" { //name undefined
		unlink_err := sumup_integration.SumupClient.Readers.DeleteReader(context.Background(), *sumup_integration.SumupAccount.MerchantProfile.MerchantCode, sumup.ReaderId(input.ReaderId))
		if unlink_err != nil {
			fmt.Printf("error while unlinking reader by id: %s\n", unlink_err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": unlink_err.Error()})
			return
		}

		if delete_err := DeleteReaderById(input.ReaderId); delete_err != nil {
			fmt.Printf("error while deleting reader by id: %s\n", delete_err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": delete_err.Error()})
		}
	} else {
		fmt.Printf("unknown error while unlinking reader\n")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "unknown error while unlinking reader"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": "success"})
}

func GetIncomingWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("incoming webhook data: %s\n", body)
	c.JSON(http.StatusOK, gin.H{"data": "success"})
}
