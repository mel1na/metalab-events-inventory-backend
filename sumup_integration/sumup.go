package sumup_integration

import (
	"context"
	"fmt"
	"metalab/events-inventory-tracker/models"
	sumup_models "metalab/events-inventory-tracker/models/sumup"
	"os"
	"time"

	"github.com/sumup/sumup-go"
)

var SumupAccount *sumup.MerchantAccount
var SumupClient *sumup.Client

func Login() {
	SumupClient = sumup.NewClient().WithAuth(os.Getenv("SUMUP_KEY"))

	account, err := SumupClient.Merchant.Get(context.Background(), sumup.GetAccountParams{})
	if err != nil {
		fmt.Printf("[ERROR] SumUp API: error getting merchant account: %s\n", err.Error())
		return
	}

	fmt.Printf("SumUp API: authorized for merchant %q (%s)\n\n", *account.MerchantProfile.MerchantCode, *account.MerchantProfile.CompanyName)
	SumupAccount = account
}

func StartReaderCheckout(ReaderId string, TotalAmount uint, Description *string) (ClientTransactionId string, Error error) {
	var returnUrl string = os.Getenv("SUMUP_RETURN_URL")
	response, checkout_err := SumupClient.Readers.CreateCheckout(context.Background(), *SumupAccount.MerchantProfile.MerchantCode, ReaderId, sumup.CreateReaderCheckoutBody{Description: Description, ReturnUrl: &returnUrl, TotalAmount: sumup.CreateReaderCheckoutAmount{Currency: "EUR", MinorUnit: 2, Value: int(TotalAmount)}})
	if checkout_err != nil {
		return "error", fmt.Errorf("[ERROR] SumUp API: error while creating reader checkout: %s\n", checkout_err.Error())
	}
	return *response.Data.ClientTransactionId, nil
}

func InitiallyCheckIfReaderIsReady(ReaderId string) (Result *sumup_models.Reader, Error error) {
	readerReady := false
	count := 5
	seconds_between := 5
	for i := 0; i <= count; i++ {
		//response, err := SumupClient.Readers.List(context.Background(), *SumupAccount.MerchantProfile.MerchantCode)
		reader, err := SumupClient.Readers.Get(context.Background(), *SumupAccount.MerchantProfile.MerchantCode, sumup.ReaderId(ReaderId), sumup.GetReaderParams{})
		if err != nil {
			fmt.Printf("error getting reader %s (interation %d/%d): %s\n", ReaderId, i, count, err.Error())
			time.Sleep(time.Second * time.Duration(seconds_between))
			continue
		}
		if reader.Status != sumup.ReaderStatusPaired {
			fmt.Printf("reader %s not ready (iteration %d/%d)\n", ReaderId, i, count)
			time.Sleep(time.Second * time.Duration(seconds_between))
			continue
		}
		fmt.Printf("reader %s returned ready\n", ReaderId)
		readerReady = true
		break
	}
	if readerReady {
		edited_reader := sumup_models.Reader{Status: sumup_models.ReaderStatusPaired}
		models.DB.Where(&sumup_models.Reader{ReaderId: sumup_models.ReaderId(ReaderId)}).Updates(edited_reader)
		fmt.Printf("reader %s is ready\n", ReaderId)
		return &edited_reader, nil
	}
	fmt.Printf("reader %s not ready after waiting %d seconds\n", ReaderId, count*seconds_between)
	return nil, fmt.Errorf("reader %s not ready after waiting %d seconds\n", ReaderId, count*seconds_between)
}

func CheckIfReaderIsReady(ReaderId string) (IsReady bool, Error error) {
	reader, err := SumupClient.Readers.Get(context.Background(), *SumupAccount.MerchantProfile.MerchantCode, sumup.ReaderId(ReaderId), sumup.GetReaderParams{})
	if err != nil {
		fmt.Printf("error getting reader %s: %s\n", ReaderId, err.Error())
		return false, err
	}
	if reader.Status != sumup.ReaderStatusPaired {
		fmt.Printf("reader %s not ready\n", ReaderId)
		return false, fmt.Errorf("reader not ready yet")
	}
	edited_reader := sumup_models.Reader{Status: sumup_models.ReaderStatusPaired}
	models.DB.Where(&sumup_models.Reader{ReaderId: sumup_models.ReaderId(ReaderId)}).Updates(edited_reader)
	fmt.Printf("reader %s returned ready\n", ReaderId)
	return true, nil
}
