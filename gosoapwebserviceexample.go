package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"text/template"
	"log"

	"github.com/clbanning/mxj"
	"github.com/mitchellh/mapstructure"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	weather, err := queryWeatherForZip("90210")
	if err != nil {
		log.Println(err)
	} else {
		spew.Dump(weather)
	}
}

type WeatherInfo struct {
	State string 
	City string 
	WeatherStationCity string
	WeatherID int 
	Description string
	Temperature string 
	RelativeHumidity string 
	Wind string
	Pressure string
	Visibility string
	WindChill string
	Remarks string
}


func queryWeatherForZip(postalCode string) (*WeatherInfo, error) {
	url := "http://wsf.cdyne.com/WeatherWS/Weather.asmx"
	client := &http.Client{}
	sRequestContent := generateRequestContent(postalCode)
	requestContent := []byte(sRequestContent)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestContent))
	if err != nil {
		return nil,err
	}

	req.Header.Add("SOAPAction", `"http://ws.cdyne.com/WeatherWS/GetCityWeatherByZIP"`)
	req.Header.Add("Content-Type", "text/xml; charset=utf-8")
	req.Header.Add("Accept", "text/xml")
	resp, err := client.Do(req)
	if err != nil {
		return nil,err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.New("Error Respose " + resp.Status)
	}
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//	sContents := string(contents)
	//	log.Println(sContents)
	m, _ := mxj.NewMapXml(contents, true)
	return convertResults(&m)
}

func generateRequestContent(postalCode string) string {
	type QueryData struct {
		PostalCode string 
	}	

	const getTemplate = `<?xml version="1.0" encoding="utf-8"?>
	<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
	  <soap:Body>
	    <GetCityWeatherByZIP xmlns="http://ws.cdyne.com/WeatherWS/">
	      <ZIP>{{.PostalCode}}</ZIP>
	    </GetCityWeatherByZIP>
	  </soap:Body>
	</soap:Envelope>`
	querydata := QueryData{PostalCode:postalCode}
	tmpl, err := template.New("getCityWeatherByZIPTemplate").Parse(getTemplate)
	if err != nil {
		panic(err)
	}
	var doc bytes.Buffer
	err = tmpl.Execute(&doc, querydata)
	if err != nil {
		panic(err)
	}
	return doc.String()
}

func convertResults(soapResponse *mxj.Map) (*WeatherInfo, error) {
	successStatus, _ := soapResponse.ValueForPath("Envelope.Body.GetCityWeatherByZIPResponse.GetCityWeatherByZIPResult.Success")
	success := successStatus.(bool)
	if !success {
		errorMessage, _ := soapResponse.ValueForPath("Envelope.Body.GetCityWeatherByZIPResponse.GetCityWeatherByZIPResult.ResponseText")
		return nil, errors.New("Error Respose " + errorMessage.(string))
	}

	weatherResult, err := soapResponse.ValueForPath("Envelope.Body.GetCityWeatherByZIPResponse.GetCityWeatherByZIPResult")
	if err != nil {
		return nil, err
	}
	var result WeatherInfo

	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &result,
		// add a DecodeHook here if you need complex Decoding of results -> DecodeHook: yourfunc,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return nil,err
	}
	if err := decoder.Decode(weatherResult); err != nil {
		return nil,err
	}
	return &result, nil
}