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

type CompanyAccounts struct {
	CompanyAccount CompanyAccount `json:"companyAccount"`
}
type CompanyAccount struct {
	Address         string   `json:"Address"`
	City            string   `json:"City"`
	CompanyName     string   `json:"CompanyName"`
	ContractID      string   `json:"ContractId"`
	Country         string   `json:"Country"`
	CrefoNumber     string   `json:"CrefoNumber"`
	CustomerID      string   `json:"CustomerId"`
	Domain          []string `json:"Domain"`
	Email           string   `json:"Email"`
	Marketplaces    []int    `json:"Marketplaces"`
	ParentAccountID int      `json:"ParentAccountId"`
	VATID           string   `json:"VATID"`
	Zip             string   `json:"Zip"`
}

//ConnectTest sets up the basics for testing
func ConnectTest(ctx context.Context) (alsocloud *Client, err error) {

	alsocloud, err = NewClient(
		"https://marketplacetest.ccpaas.net",
		os.Getenv("ALSO_CLOUDUSER"),
		os.Getenv("ALSO_CLOUDPW"),
		&Options{
			Timeout: 30},
	)
	return alsocloud, err
}

//TestConnect_Error tests incorrect URL format
func TestConnect_Error(t *testing.T) {
	_, err := NewClient(
		"google.ch", // Invalid URL without scheme
		"demouser",
		"1234",
		nil,
	)

	//Check error; should not be nil
	if err == nil {
		t.Errorf("Expected Error. Got '%v'", err)
	}
}
func TestClient_GetCompanies(t *testing.T) {
	ctx := context.Background()
	alsorest, err := ConnectTest(ctx)
	if err != nil {
		t.Skipf("Skipping test - connection failed: %v", err)
	}

	var data map[string]interface{} = map[string]interface{}{
		"parentAccountId": 523525,
	}

	resp, _, statuscode, err := alsorest.Post(ctx, "GetCompany", data)
	if resp != nil {
		defer resp.Close()
	}

	//Check status code; Should be 200
	if statuscode != 200 {
		t.Errorf("Expected HTTP Status Code 200. Got '%v'", statuscode)
	}

	//Check error; Should be nil
	if err != nil {
		t.Errorf("Expected no error. Got '%v'", err)
	}

	//Validate session
	valid := alsorest.Validate(ctx)

	//Check if session is still valid - should be
	if !valid {
		t.Errorf("Session should be still valid but is '%v'", valid)
	}
}

func Test_errorFormatter(t *testing.T) {
	ctx := context.Background()
	client, err := NewClient(Debug, "demo", "1234", nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	xmlError, err := os.Open("sampleError.xml")
	if err != nil {
		t.Fatalf("Can't open sampleXml: '%v'", err)
	}
	defer xmlError.Close()

	err = errorFormatterPx(ctx, client, 400, xmlError)

	if err == nil || err.Error() == "" {
		t.Errorf("Error shouldn't be nil or empty: '%v'", err)
	}
}

func TestClient_Timeout(t *testing.T) {
	_, err := NewClient(
		"https://marketplacetest.ccpaas.net",
		"test@example.com",
		"password",
		&Options{Timeout: 1}, // 1 second timeout
	)

	if err != nil {
		t.Errorf("Expected no error creating client with timeout. Got '%v'", err)
	}
}

//func TestClient_CreateCompany(t *testing.T) {
//	ctx := context.Background()
//	alsorest, err := ConnectTest(ctx)
//
//	data := CompanyAccount{
//		Address:         "Demoaddress",
//		City:            "Zurich",
//		CompanyName:     "Demo Muster Debug Company",
//		ContractID:      "",
//		Country:         "Switzerland",
//		CrefoNumber:     "",
//		CustomerID:      "",
//		Domain:          nil,
//		Email:           "",
//		Marketplaces:    nil,
//		ParentAccountID: 0,
//		VATID:           "",
//		Zip:             "8000",
//	}
//	_, _, statuscode, err := alsorest.Post(ctx, "CreateCompany", data)
//
//
//
//}
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
