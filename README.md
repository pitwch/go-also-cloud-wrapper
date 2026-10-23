[![GitHub release](https://img.shields.io/github/tag/pitwch/go-also-cloud-wrapper.svg?label=latest&colorB=orange)](https://github.com/pitwch/go-also-cloud-wrapper/releases/latest)
[![CI](https://github.com/pitwch/go-also-cloud-wrapper/actions/workflows/ci.yml/badge.svg)](https://github.com/pitwch/go-also-cloud-wrapper/actions/workflows/ci.yml)
[![Release](https://github.com/pitwch/go-also-cloud-wrapper/actions/workflows/release.yml/badge.svg)](https://github.com/pitwch/go-also-cloud-wrapper/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/pitwch/go-also-cloud-wrapper)](https://goreportcard.com/report/github.com/pitwch/go-also-cloud-wrapper)
[![GoDoc](https://godoc.org/github.com/pitwch/go-also-cloud-wrapper/alsocloud?status.svg)](https://godoc.org/github.com/pitwch/go-also-cloud-wrapper/alsocloud)
[![codecov](https://codecov.io/gh/pitwch/go-also-cloud-wrapper/branch/master/graph/badge.svg)](https://codecov.io/gh/pitwch/go-also-cloud-wrapper)
[![GitHub license](https://img.shields.io/github/license/pitwch/go-also-cloud-wrapper.svg)](https://github.com/pitwch/go-also-cloud-wrapper/blob/master/LICENSE)

# ALSO Cloud Marketplace API Golang Wrapper

Golang Wrapper for ALSO Cloud Marketplace API


![alt text](https://raw.githubusercontent.com/pitwch/go-also-cloud-wrapper/master/_assets/img/gopher-also.png "ALSO Cloud Marketplace API Golang Wrapper")

Swagger Specs for ALSO Cloud Marketplace Simple API are here:
https://app.swaggerhub.com/apis/MarketplaceSimpleAPI/MarketplaceSimpleAPI/1.0.0

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
| options       | &also.Options{Timeout: 30 * time.Second}  | `*also.Options` | Options (see Chapter `Options`)       |


Example:

```golang
import (
  also "github.com/pitwch/go-also-cloud-wrapper/alsocloud"
)

var alsocloud, err = also.NewClient(
	"https://marketplace.also.ch",
	"demo@example.com",
	"1234",
	&also.Options{Log: true, Timeout: 30 * time.Second},
)
```


### Options

Options are **optional**:

| Option        | Example                              | Note                                                         |
|---------------|--------------------------------------|--------------------------------------------------------------|
| APIPrefix     | /SimpleAPI/SimpleAPIService.svc/rest | API - Prefix; Default = /SimpleAPI/SimpleAPIService.svc/rest |
| LoginEndpoint | GetSessionToken                      | Endpoint for Login; Default = GetSessionToken                |
| UserAgent     | go-also-cloud-wrapper                | User Agent; Default = go-also-cloud-wrapper \<version\>      |
| Timeout       | 30 * time.Second                     | HTTP timeout (`Timeout: 30` still works as legacy seconds)   |
| VerifySSL     | true                                 | Kept for compatibility; TLS is always verified by default. Custom TLS via `HTTPClient`. |
| Log           | true                                 | Activates Log Output; Default = false                        |
| HTTPClient    | urlfetch.Client(ctx) <sup>[1](#urlfetch)</sup> | Custom HTTP client; Default = package transport              |
| Batchsize     | 200                                  | Reserved for future batch helpers; Default = 200             |

<a name="urlfetch">1</a>: urlfetch is used for Google App Engine Standard Instances but can be replaced by any HTTP-Client



#### Methods

The wrapper supports all HTTP methods for interacting with the ALSO Cloud Marketplace API:

**Available Methods:**
- `Post(ctx, endpoint, data)` - POST requests (creating resources)
- `Get(ctx, endpoint, params)` - GET requests (fetching data with query parameters)
- `Put(ctx, endpoint, data)` - PUT requests (updating resources)
- `Patch(ctx, endpoint, data)` - PATCH requests (partial updates)
- `Delete(ctx, endpoint, data)` - DELETE requests (removing resources)

**Parameters:**

| Parameter  | Type          | Note                                                          |
|------------|---------------|---------------------------------------------------------------|
| ctx        | `context.Context` | Context for request cancellation and deadlines            |
| endpoint   | `string`      | Endpoint ALSO Cloud Marketplace; f.ex. GetCompany, GetUser... |
| data       | `interface{}` | Data (automatic conversion to JSON)                           |
| params     | `url.Values`  | URL query parameters (for GET requests)                       |


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
	&also.Options{Log: true, Timeout: 30 * time.Second},
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

### Testing

The tests require valid ALSO Cloud credentials to run. Set the following environment variables:

**Local Testing:**

```bash
# Windows (PowerShell)
$env:ALSO_CLOUDUSER="your-username@example.com"
$env:ALSO_CLOUDPW="your-password"
go test ./...

# Linux/macOS
export ALSO_CLOUDUSER="your-username@example.com"
export ALSO_CLOUDPW="your-password"
go test ./...
```

**GitHub Actions:**

For CI/CD pipelines, store credentials as GitHub Secrets:

1. Navigate to **Settings** → **Secrets and variables** → **Actions**
2. Add two repository secrets:
   - `ALSO_CLOUDUSER`: Your ALSO Cloud username
   - `ALSO_CLOUDPW`: Your ALSO Cloud password

The GitHub workflow will automatically use these secrets as environment variables during test execution.

---

For a full list of options check [GoDoc](https://godoc.org/github.com/pitwch/go-also-cloud-wrapper/alsocloud).
