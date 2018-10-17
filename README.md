[![Go Report Card](https://goreportcard.com/badge/github.com/pitwch/go-also-cloud-wrapper)](https://goreportcard.com/report/github.com/pitwch/go-also-cloud-wrapper)
[![GoDoc](https://godoc.org/github.com/pitwch/go-also-cloud-wrapper/alsocloud?status.svg)](https://godoc.org/github.com/pitwch/go-also-cloud-wrapper/alsocloud)
[![Build Status](https://travis-ci.org/pitwch/go-also-cloud-wrapper.svg?branch=master)](https://travis-ci.org/pitwch/go-also-cloud-wrapper)
[![codecov](https://codecov.io/gh/pitwch/go-also-cloud-wrapper/branch/master/graph/badge.svg)](https://codecov.io/gh/pitwch/go-also-cloud-wrapper)

# ALSO Cloud Marketplace API Golang Wrapper

Golang Wrapper for the ALSO Cloud Marketplace API


![alt text](https://raw.githubusercontent.com/pitwch/go-also-cloud-wrapper/master/assets/img/gopher-also.png "ALSO Cloud Marketplace API Golang Wrapper")

Swagger Specs for ALSO Cloud Marketplace Simple API are here:
https://app.swaggerhub.com/apis/Marketplace_SimpleAPI/Mareketplace_SimpleAPI/1.0.0

### Installation

```bash
$ go get github.com/pitwch/go-also-cloud-wrapper/alsocloud
```


#### Configuration


| Configuration | Examples                    | Type            | Note                                  |
|---------------|-----------------------------|-----------------|---------------------------------------|
| Marketplace   | https://marketplace.also.ch | `string`        | URL Marketplace (Country specific)    |
| apiUser       | demo@example.com            | `string`        | Username for Cloud Marketplace        |
| apiPassword   | 1234                        | `string`        | Password for Cloud Marketplace        |
| options       | &also.Options{Timeout: 30}  | `*also.Options` | Options (see Chapter `Options`)       |


Example:

```golang
import (
  also "github.com/pitwch/go-also-cloud-wrapper/alsocloud"
)

var alsocloud, err = also.NewClient(
	"https://marketplace.also.ch",
	"demo@example.com",
	"1234",
	&px.Options{Log: true, Timeout: 30},
)
```


### Options

Options are **optional**:

| Option        | Example                              | Note                                                         |
|---------------|--------------------------------------|--------------------------------------------------------------|
| APIPrefix     | /SimpleAPI/SimpleAPIService.svc/rest | API - Prefix; Default = /SimpleAPI/SimpleAPIService.svc/rest |
| LoginEndpoint | GetSessionToken                      | Endpoint for Login; Default = GetSessionToken                |
| UserAgent     | go-also-cloud-wrapper                | User Agent; Default = go-also-cloud-wrapper                  |
| Timeout       | 15                                   | Timeout in seconds                                           |
| VerifySSL     | true                                 | Check if SSL is valid                                        |
| Log           | true                                 | Activates Log Output; Default = false                        |
| Client        | urlfetch.Client(ctx) <sup>[1](#urlfetch)</sup>                | HTTP-Client; Default = http.DefaultClient                    |

<a name="urlfetch">1</a>: urlfetch is used for Google App Engine Standard Instances but can be replaced by any HTTP-Client



#### Methods


| Parameter  | Typ           | Note                                                          |
|------------|---------------|---------------------------------------------------------------|
| endpoint   | `string`      | Endpoint ALSO Cloud Marketplace; f.ex. GetCompany, GetUser... |
| data       | `interface{}` | Date (automatic conversion to JSON)                           |
| parameters | `url.Values`  | Parameters                                                    |


#### Full Example

**Get Company**

```golang
import (
  also "github.com/pitwch/go-also-cloud-wrapper/alsocloud"
)

//Create client
var alsocloud, err = also.NewClient(
	alsocloud.Switzerland,			//Using predefined Constant Switzerland.  
	"demo@example.com",			//Plain URL would be possible "https://marketplace.also.ch"
	"1234",
	&px.Options{Log: true, Timeout: 30},
)

//Create context
ctx := context.Background()

//Query
res, _, _, err := alsocloud.Post(ctx, "GetCompany", nil)

	buf := new(bytes.Buffer)
	buf.ReadFrom(res)
	resp := buf.String()
	fmt.Printf(resp, err)
	defer res.Close()
	
	//returns {"ParentAccountId": 1234,"AccountId": 1234,"AccountState": "Active","CompanyName": "Demo"...

```

For a full list of options check [GoDoc](https://godoc.org/github.com/pitwch/go-also-cloud-wrapper/alsocloud).
