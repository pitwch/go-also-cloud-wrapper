package alsocloud

import (
	"context"
	"os"
	"testing"
)

type Invoice struct {
	CompanyName      string   `json:"CompanyName"`
	CompanyAccountID int      `json:"CompanyAccountId"`
	CompanyVatID     string   `json:"CompanyVatId"`
	BillingInterval  string   `json:"BillingInterval"`
	Charges          []Charge `json:"Charges"`
}

type Charge struct {
	ServiceName              string  `json:"ServiceName"`
	ServiceID                string  `json:"ServiceId"`
	AccountID                int     `json:"AccountId"`
	PriceableItemDescription string  `json:"PriceableItemDescription"`
	MaterialNumber           string  `json:"MaterialNumber"`
	BillableParameter        string  `json:"BillableParameter"`
	ContractID               string  `json:"ContractId"`
	BillingStartDate         string  `json:"BillingStartDate"`
	ActualChargeInterval     string  `json:"ActualChargeInterval"`
	Costs                    float64 `json:"Costs"`
	SalesPrice               float64 `json:"SalesPrice"`
	Currency                 string  `json:"Currency"`
}

//Test_EnvVars Checks if encrypted EnvVars aren't empty
func Test_EnvVars(t *testing.T) {
	if os.Getenv("ALSO_CLOUDUSER") == "" {
		t.Errorf("Environment Var %v is empty", "ALSO_CLOUDUSER")

	}
	if os.Getenv("ALSO_CLOUDPASSWORD") == "" {
		t.Errorf("Environment Var %v is empty", "ALSO_CLOUDPASSWORD")

	}

}

//ConnectTest sets up the basics for testing
func ConnectTest(ctx context.Context) (alsocloud *Client, err error) {

	alsocloud, err = NewClient(
		"https://marketplacetest.ccpaas.net",
		os.Getenv("ALSO_CLOUDUSER"),
		os.Getenv("ALSO_CLOUDPASSWORD"),
		&Options{
			Timeout: 30},
	)
	return alsocloud, err
}
func TestClient_GetCompanies(t *testing.T) {

	ctx := context.Background()
	alsorest, err := ConnectTest(ctx)
	var data map[string]interface{} = map[string]interface{}{
		"parentAccountId": 523525,
	}

	_, _, statuscode, err := alsorest.Post(ctx, "GetCompany", data)

	//Check status code; Should be 201
	if statuscode != 200 {
		t.Errorf("Expected HTTP Status Code 200. Got '%v'", statuscode)
	}

	//Check error; Should be nil
	if err != nil {
		t.Errorf("Expected no error. Got '%v'", err)
	}

}

//func TestClient_GetLatestInvoice(t *testing.T) {
//	ctx := context.Background()
//	alsorest, err := ConnectTest(ctx)
//	var data map[string]interface{} = map[string]interface{}{
//		"parentAccountId": 523525,
//	}
//
//	_, _, statuscode, err := alsorest.Post(ctx, "GetLatestInvoices", data)
//
//	//Check status code; Should be 201
//	if statuscode != 200 {
//		t.Errorf("Expected HTTP Status Code 200. Got '%v'", statuscode)
//	}
//
//	//Check error; Should be nil
//	if err != nil {
//		t.Errorf("Expected no error. Got '%v'", err)
//	}
//
//}
