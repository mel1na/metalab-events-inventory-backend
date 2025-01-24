package sumup_integration

import (
	"context"
	"fmt"
	"os"

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

func StartReaderCheckout(ReaderId string, TotalAmount uint) (ClientTransactionId string, Error error) {
	var returnUrl string = os.Getenv("SUMUP_RETURN_URL")
	response, checkout_err := SumupClient.Readers.CreateCheckout(context.Background(), *SumupAccount.MerchantProfile.MerchantCode, ReaderId, sumup.CreateReaderCheckoutBody{ReturnUrl: &returnUrl, TotalAmount: sumup.CreateReaderCheckoutAmount{Currency: "EUR", MinorUnit: 2, Value: int(TotalAmount)}})
	if checkout_err != nil {
		return "error", fmt.Errorf("[ERROR] SumUp API: error while creating reader checkout: %s\n", checkout_err.Error())
	}
	return *response.Data.ClientTransactionId, nil
}
