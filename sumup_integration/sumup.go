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
		fmt.Printf("SumUp API: get merchant account: %s\n", err.Error())
		return
	}

	fmt.Printf("SumUp API: authorized for merchant %q\n\n", *account.MerchantProfile.MerchantCode)
	SumupAccount = account
}

func StartReaderCheckout(ReaderId string, TotalAmount uint) (ClientTransactionId string, Error error) {
	response, checkout_err := SumupClient.Readers.CreateCheckout(context.Background(), *SumupAccount.MerchantProfile.MerchantCode, ReaderId, sumup.CreateReaderCheckoutBody{TotalAmount: sumup.CreateReaderCheckoutAmount{Currency: "EUR", MinorUnit: 2, Value: int(TotalAmount)}})
	if checkout_err != nil {
		return "error", fmt.Errorf("error while creating reader checkout: %s\n", checkout_err.Error())
	}
	return *response.Data.ClientTransactionId, nil
}
