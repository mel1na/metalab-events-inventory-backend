package sumup_integration

import (
	"context"
	"fmt"
	"metalab/events-inventory-tracker/models"
	sumup_models "metalab/events-inventory-tracker/models/sumup"
	"os"
	"time"

	"github.com/sumup/sumup-go"
	"gorm.io/gorm"
)

var SumupAccount *sumup.MerchantAccount
var SumupClient *sumup.Client

func Login() {
	SumupClient = sumup.NewClient().WithAuth(os.Getenv("SUMUP_KEY"))

	account, err := SumupClient.Merchant.Get(context.Background(), sumup.GetAccountParams{})
	if err != nil {
		fmt.Printf("[ERROR] SumUp API: Error getting merchant account: %s\n", err.Error())
		return
	}

	fmt.Printf("SumUp API: Authorized for merchant %q (%s)\n\n", *account.MerchantProfile.MerchantCode, *account.MerchantProfile.CompanyName)
	SumupAccount = account
}

func InitAPIReaders() {
	response, err := SumupClient.Readers.List(context.Background(), *SumupAccount.MerchantProfile.MerchantCode)
	if err != nil {
		fmt.Printf("Error fetching readers from SumUp API: %s\n", err.Error())
		return
	}

	var readers []sumup_models.Reader
	models.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&readers)

	// lookup if readers are in db by reader id, create only non added ones.
	readers_count := 0
	for _, v := range response.Items {
		api_reader := sumup_models.Reader{ReaderId: sumup_models.ReaderId(v.Id), Name: sumup_models.ReaderName(v.Name), Status: sumup_models.ReaderStatus(v.Status), Device: sumup_models.ReaderDevice{Identifier: v.Device.Identifier, Model: sumup_models.ReaderDeviceModel(v.Device.Model)}, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
		models.DB.Create(&api_reader)
		readers_count++
	}
	fmt.Printf("Intitialized %d reader(s) from API.\n", readers_count)
}

func StartReaderCheckout(ReaderId string, TotalAmount uint, Description *string) (ClientTransactionId string, Error error) {
	var returnUrl string = os.Getenv("SUMUP_RETURN_URL")
	response, checkout_err := SumupClient.Readers.CreateCheckout(context.Background(), *SumupAccount.MerchantProfile.MerchantCode, ReaderId, sumup.CreateReaderCheckoutBody{Description: Description, ReturnUrl: &returnUrl, TotalAmount: sumup.CreateReaderCheckoutAmount{Currency: "EUR", MinorUnit: 2, Value: int(TotalAmount)}})
	if checkout_err != nil {
		return "error", fmt.Errorf("[ERROR] SumUp API: Error while creating reader checkout: %s", checkout_err.Error())
	}
	return *response.Data.ClientTransactionId, nil
}

func InitiallyCheckIfReaderIsReady(ReaderId string) (Result *sumup_models.Reader, Error error) {
	readerReady := false
	count := 5
	seconds_between := 5
	for i := 0; i <= count; i++ {
		time.Sleep(time.Second * time.Duration(seconds_between))
		//response, err := SumupClient.Readers.List(context.Background(), *SumupAccount.MerchantProfile.MerchantCode)
		reader, err := SumupClient.Readers.Get(context.Background(), *SumupAccount.MerchantProfile.MerchantCode, sumup.ReaderId(ReaderId), sumup.GetReaderParams{})
		if err != nil {
			fmt.Printf("Error getting reader %s (Iteration %d/%d): %s\n", ReaderId, i, count, err.Error())
			continue
		}
		if reader.Status != sumup.ReaderStatusPaired {
			fmt.Printf("Reader %s not ready (Iteration %d/%d)\n", ReaderId, i, count)
			continue
		}
		fmt.Printf("Reader %s returned ready\n", ReaderId)
		readerReady = true
		break
	}
	if readerReady {
		edited_reader := sumup_models.Reader{Status: sumup_models.ReaderStatusPaired}
		models.DB.Where(&sumup_models.Reader{ReaderId: sumup_models.ReaderId(ReaderId)}).Updates(edited_reader)
		fmt.Printf("Reader %s is ready\n", ReaderId)
		return &edited_reader, nil
	}
	fmt.Printf("Reader %s not ready after waiting %d seconds\n", ReaderId, count*seconds_between)
	return nil, fmt.Errorf("reader %s not ready after waiting %d seconds", ReaderId, count*seconds_between)
}

func CheckIfReaderIsReady(ReaderId string) (IsReady bool, Error error) {
	reader, err := SumupClient.Readers.Get(context.Background(), *SumupAccount.MerchantProfile.MerchantCode, sumup.ReaderId(ReaderId), sumup.GetReaderParams{})
	if err != nil {
		fmt.Printf("Error getting reader %s: %s\n", ReaderId, err.Error())
		return false, err
	}
	if reader.Status != sumup.ReaderStatusPaired {
		fmt.Printf("Reader %s not ready\n", ReaderId)
		return false, fmt.Errorf("reader not ready yet")
	}
	edited_reader := sumup_models.Reader{Status: sumup_models.ReaderStatusPaired}
	models.DB.Where(&sumup_models.Reader{ReaderId: sumup_models.ReaderId(ReaderId)}).Updates(edited_reader)
	fmt.Printf("Reader %s returned ready\n", ReaderId)
	return true, nil
}
