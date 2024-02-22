package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Attribute struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

// Define the Traits structure
type Traits struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

// Define the ChangeFormatContactStruct structure
type ChangeFormatContactStruct struct {
	Event           string               `json:"event"`
	EventType       string               `json:"event_type"`
	AppID           string               `json:"app_id"`
	UserID          string               `json:"user_id"`
	MessageID       string               `json:"message_id"`
	PageTitle       string               `json:"page_title"`
	PageURL         string               `json:"page_url"`
	BrowserLanguage string               `json:"browser_language"`
	ScreenSize      string               `json:"screen_size"`
	Attributes      map[string]Attribute `json:"attributes"`
	Traits          map[string]Traits    `json:"traits"`
}

func main() {
	//create a new router
	router := mux.NewRouter()

	InitializeRoutes(router)
	fmt.Printf("\nListening to port localhost:9060\n")
	err := http.ListenAndServe(":9009", router)
	if err != nil {
		fmt.Print(err)
	}
}

// create a success response
func SuccessResp(w http.ResponseWriter, statusCode int, data interface{}) {
	// Set the HTTP status code
	w.WriteHeader(statusCode)

	//create a response struct and get the values from paramer and pass the struct
	c := struct {
		Status bool        `json:"status"`
		Data   interface{} `json:"data"`
	}{Status: true, Data: data}

	// convert struct into JSON format
	b, err := json.MarshalIndent(&c, "", "\t")
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
	// encode the JSON response and write the response writer
	err = json.NewEncoder(w).Encode(json.RawMessage(string(b)))
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

// create a failure response
func FailureResp(w http.ResponseWriter, statusCode int, failureMsg interface{}) {
	// Set the HTTP status code
	w.WriteHeader(statusCode)

	//create a response struct and get the values from paramer and pass the struct
	c := struct {
		Status bool        `json:"status"`
		Data   interface{} `json:"data"`
		Error  interface{} `json:"error"`
	}{Status: false, Data: nil, Error: failureMsg}

	// convert struct into JSON format
	b, err := json.MarshalIndent(&c, "", "\t")
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}

	// encode the JSON response and write the response writer
	err = json.NewEncoder(w).Encode(json.RawMessage(string(b)))
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

// SetMiddlewareJSON is used to set the content-type header
func SetMiddlewareJSON(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next(w, r)
	}
}

// InitializeRoutes is used to route the function
func InitializeRoutes(router *mux.Router) {
	router.HandleFunc("/contact-form",
		SetMiddlewareJSON((ContactForm))).Methods(http.MethodPost)
}

func ContactForm(w http.ResponseWriter, r *http.Request) {
	//get the body request
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error occurs on unmarshal the body")
		FailureResp(w, http.StatusUnprocessableEntity, err)
		return
	}

	// Create the empty struct variable for unmarshal the jsonData value to empty struct
	var contactStruct map[string]interface{}

	// Unmarshal the body to an empty struct
	err = json.Unmarshal(body, &contactStruct)
	if err != nil {
		fmt.Println("error occurs on unmarshal the body")
		FailureResp(w, http.StatusUnprocessableEntity, err)
		return
	}

	// Create a new Channel
	ch := make(chan ChangeFormatContactStruct)

	// Start the do routing and send the result through the channel
	go GoWorker(contactStruct, ch)

	// Wait for response from the channel
	resp := <-ch
	fmt.Println(resp)

	//emcode z
	convertedResp, err := json.Marshal(resp)
	if err != nil {
		FailureResp(w, http.StatusUnprocessableEntity, err)
		return
	}

	//create a weebhook url
	webHookUrl := "https://webhook.site/58d47518-6286-4022-87e7-816049dbacc6"

	//send the response in webhook url
	WebHookResp, err := http.Post(webHookUrl, "application/json", bytes.NewBuffer(convertedResp))
	if err != nil {
		FailureResp(w, http.StatusUnprocessableEntity, err)
		return
	}
	defer WebHookResp.Body.Close()
	// Send the response to http server
	SuccessResp(w, http.StatusAccepted, resp)
}

func GoWorker(contactStruct map[string]interface{}, ch chan<- ChangeFormatContactStruct) {
	// Create the empty attributes and traits struct
	attributes := make(map[string]Attribute)
	traits := make(map[string]Traits)

	for key, _ := range contactStruct {
		//check if the key is atrk or uatrk
		if strings.HasPrefix(key, "atrk") {
			// replace the value, type and key value from key
			valueKey := strings.Replace(key, "atrk", "atrv", 1)
			typeKey := strings.Replace(key, "atrk", "atrt", 1)
			keyKey := strings.Replace(key, "atrk", "atrk", 1)
			//get the value, type, and key of attribute
			valueValue, exists := contactStruct[valueKey].(string)
			if !exists {
				continue
			}

			valueType, exists := contactStruct[typeKey].(string)
			if !exists {
				continue
			}

			valueKay, exists := contactStruct[keyKey].(string)
			if !exists {
				continue
			}

			// set the value, type, and key values in attrubute struct
			attributes[valueKay] = Attribute{
				Value: valueValue,
				Type:  valueType,
			}
		} else if strings.HasPrefix(key, "uatrk") {
			// replace the value, type and key value from key
			valueKey := strings.Replace(key, "uatrk", "uatrv", 1)
			typeKey := strings.Replace(key, "uatrk", "uatrt", 1)
			keyKey := strings.Replace(key, "uatrk", "uatrk", 1)
			//get the value, type, and key of trial
			valueValue, exists := contactStruct[valueKey].(string)
			if !exists {
				continue
			}

			valueType, exists := contactStruct[typeKey].(string)
			if !exists {
				continue
			}

			valueKay, exists := contactStruct[keyKey].(string)
			if !exists {
				continue
			}
			// set the value, type, and key values in attrubute trial
			traits[valueKay] = Traits{
				Value: valueValue,
				Type:  valueType,
			}
		}
	}

	// Create a ChangeFormatContactStruct instance
	changeFormatStruct := ChangeFormatContactStruct{
		Event:           contactStruct["ev"].(string),
		EventType:       contactStruct["et"].(string),
		AppID:           contactStruct["id"].(string),
		UserID:          contactStruct["uid"].(string),
		MessageID:       contactStruct["mid"].(string),
		PageTitle:       contactStruct["t"].(string),
		PageURL:         contactStruct["p"].(string),
		BrowserLanguage: contactStruct["l"].(string),
		ScreenSize:      contactStruct["sc"].(string),
		Attributes:      attributes,
		Traits:          traits,
	}

	// Send the ChangeFormatContactStruct struct through the channel
	ch <- changeFormatStruct
}
