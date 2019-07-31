package authzadapter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// {"quota":10,"usage":0}
type QuotaUsage struct{
	Quota int `json:"quota"`
	Usage int `json:"usage"`
}

// policy example: {
//	"quotaUsageThreshold": 50,
//	"allowPriority": 10
//}
type EntityProperty struct{
	Policy PolicyDef
	PartnerSubscriptionId string
}
type PolicyDef struct {
	QuotaUsageThreshold int `json:"quota"`
	AllowPriority int `json:"allowPriority"`
}

type QuotaType int

const (
	PSS = iota
	ASS
)

func CheckQuotaPolicy(clientId string, priority int) error  {
	//base64:appId@spId:appPassword
	log.Printf("Check quota policy with clientId:%s priority %d", clientId, priority)
	log.Println()
	entityProperty := getEntityProperty(clientId)
	// todo: how to get appSubId and it's quota usage???
	//currentUsage, err := getQuotaUsageByClientId(clientId)
	currentUsage, err := getSpssQuotaUsage(entityProperty.PartnerSubscriptionId)
	if err != nil {
		log.Fatalln("get quota usage error: %v", err)
		return err
	}
	if currentUsage == 100 {
		log.Fatalln("Quota Exhausted!!!")
		return errors.New("Quota Exhausted!!!")
	}

	quotaUsageThreshold := entityProperty.Policy.QuotaUsageThreshold
	allowPriority := entityProperty.Policy.AllowPriority

	if currentUsage >= quotaUsageThreshold {
		if priority < allowPriority{
			errMsg := fmt.Sprintf("Request priority shall be higher than %d, Reject!!!", allowPriority)
			log.Fatalln(errMsg)
			return errors.New(errMsg)
		}
	}
	return nil
}

// todo: expose quota property and scId in params
// todo: only get partnerSubscriptionId in common service
func getQuotaUsageByClientId(clientId string) (usage int, err error) {
	spIdAndAppId := strings.Split(clientId, "@")
	var appId = spIdAndAppId[0]
	var spId = ""
	if len(spIdAndAppId) == 2 {
		spId = spIdAndAppId[1]
	}
	appSubId := appId + "-ApiProxy"
	appUsage, err := getUsagePercent(QuotaType(ASS), appSubId, "QuotaProperty")
	if err != nil {
		log.Fatalln("get application level quota usage error", err)
		return 0, err
	}
	if IsEmpty(spId) {
		return appUsage, nil
	} else {
		spSubId := spId + "-ApiProxy"
		spUsage, err := getUsagePercent(QuotaType(PSS), spSubId, "QuotaProperty")
		if err != nil {
			log.Fatalln("get partner level quota usage error", err)
			return 0, err
		}
		if appUsage > spUsage {
			return appUsage, nil
		}
		return spUsage, nil
	}
}

func getSpssQuotaUsage(spSubId string) (usage int, err error) {
	spUsage, err := getUsagePercent(QuotaType(PSS), spSubId, "APIGWQuota")
	if err != nil {
		log.Fatalln("get partner level quota usage error", err)
		return 0, err
	}
	return spUsage, nil
}

func ConsumeQuotaByClientId(clientId string) {
	spIdAndAppId := strings.Split(clientId, "@")
	var appId = spIdAndAppId[0]
	consumeQuota(appId, "QuotaProperty", 1)
}
//todo: sc is hard code now
func getEntityProperty(clientId string) *EntityProperty {
	spIdAndAppId := strings.Split(clientId, "@")
	var appId = spIdAndAppId[0]
	client := &http.Client{}

	req, err := http.NewRequest("GET", "http://localhost:27120/commonservice/propertyaccess/v1/service_subscription", nil)
	if err != nil {
		log.Fatalln(err)
		return new(EntityProperty)
	}
	params := req.URL.Query()
	params.Add("service_capability", "APIGW")
	params.Add("application", appId)

	req.URL.RawQuery = params.Encode()
	log.Println("get entity property url: " +req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
		return new(EntityProperty)
	}
	statusCode := resp.StatusCode
	if statusCode != 200{
		log.Println("get property failure")
		return new(EntityProperty)
	} else {
		log.Println("get property success")

		dataByte, err:= ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}
		data := string(dataByte)
		log.Println("property response data:", data)
		// remove some chars since it's not a valid response in json format
		data =strings.ReplaceAll(data, `"{`, `{`)
		data = strings.ReplaceAll(data, `}"`, `}`)
		data = strings.ReplaceAll(data, `\`, ``)
		log.Println("parsed property: ", data)

		//err = json.Unmarshal(dataByte, &entityProperty)
		var entityProperty EntityProperty
		err = json.Unmarshal([]byte(data), &entityProperty)
		if err!= nil {
			log.Println(err)
			os.Exit(1)
		}
		log.Println("get entity property: ", entityProperty)
		return &entityProperty
	}

}

func getUsagePercent(queryType QuotaType, subId string, quotaProperty string) (usage int, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:27120/commonservice/quota/v1/query", nil)
	if err != nil {
		log.Fatalln(err)
		return 0, err
	}
	params := req.URL.Query()
	switch queryType {
	case PSS:
		params.Add("partner_subscription", subId)
	case ASS:
		params.Add("application_subscription", subId)
	}

	params.Add("quota_name", quotaProperty)

	req.URL.RawQuery = params.Encode()
	log.Println("query quota url: " + req.URL.String())

	resp, err := client.Do(req)
	if err!= nil {
		log.Println(err)
		return 0, err
	}
	statusCode := resp.StatusCode
	if statusCode != 200{
		log.Println("query quota error")
		return 0, errors.New("Query quota error")
	} else {
		log.Println("query quota success")

		dataByte, err:= ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
			os.Exit(1)
		}
		data := string(dataByte)
		log.Println("quota usage response:", data)
		var quotaUsage QuotaUsage

		err = json.Unmarshal(dataByte, &quotaUsage)
		if err!= nil {
			log.Println(err)
			os.Exit(1)
		}
		totalQuota := quotaUsage.Quota
		currentUsage := quotaUsage.Usage

		var usagePercent int
		usagePercent = currentUsage/totalQuota * 100
		fmt.Println("current quota usage ", usagePercent)
		return usagePercent, nil
	}
}

func consumeQuota(appId string, quotaProperty string, amount int) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://localhost:27120/commonservice/quota/v1/consume", nil)
	if err != nil {
		log.Fatalln(err)
		return
	}
	params := req.URL.Query()
	params.Add("application", appId)
	params.Add("service_capability", "ApiProxy")
	params.Add("quota_name", quotaProperty)
	params.Add("amount", strconv.Itoa(amount))

	req.URL.RawQuery = params.Encode()
	log.Println(req.URL.String())

	client.Do(req)
}


func IsEmpty(str string) bool  {
	return str == "" || len(str) == 0
}

func example()  {
	CheckQuotaPolicy("app1", 1)
	consumeQuota("app1", "APIGWQuota", 8)
	CheckQuotaPolicy("app1", 11)
}
