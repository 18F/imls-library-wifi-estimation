package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"

	"gsa.gov/18f/config"
)

// A package-local counter.
// The first thing we do is post an event. This will return a "magic index"
// or a foreign key, that we will use in our post of the data. This associates
// every piece of data entered with a session, and indexes the post in that session.
// That way, we can say "this set of data was entry 293 of session ABC."
// If it isn't an event object, we won't get a magic_index back, and it will
// be returned as -1. Hopefully, we'll be ignoring it in those cases...
var magic_index int = 0

var slash_warned bool = false

func PostJSON(cfg *config.Config, uri string, data []map[string]string) (int, error) {

	tok := cfg.Auth.Token
	matched, _ := regexp.MatchString(".*/$", uri)
	if !slash_warned && !matched {
		slash_warned = true
		log.Println("WARNING: api.data.gov wants a trailing slash on URIs")
	}

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	// Lets not send too much data at once. So, we'll walk through the data array in steps of 20 elements.
	chunkSize := 20
	var divided [][]map[string]string
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}

		divided = append(divided, data[i:end])
	}

	// FIXME: If there is no data, throw an event that logs there was no data.

	// Now that the incoming array has been chopped up into subarrays of length chunkSize,
	// lets send those out into the world.
	for _, arr := range divided {
		var reqBody []byte
		var err error

		// First, try marshalling the data.
		// We have to give up if this doesn't work.
		reqBody, err = json.Marshal(arr)
		if err != nil {
			return -1, errors.New("postjson: unable to marshal post of data to JSON")
		}

		// Next, it's time to create a request object. Again, fail if it doesn't work.
		req, err := http.NewRequest("POST", uri, bytes.NewBuffer(reqBody))
		if err != nil {
			return -1, errors.New("postjson: unable to construct request for data POST")
		}

		// Lets set some headers. Perhaps we should quit here... but, we'll try and keep going
		// if anything fails.
		req.Header.Set("Content-type", "application/json")
		req.Header.Set("X-Api-Key", tok)

		// MAKE THE REQUEST
		resp, err := client.Do(req)
		// If there was an error in the post, log it, and exit the function.
		if err != nil {
			message := fmt.Sprintf("postjson: failure in client attempt to POST to %v", uri)
			log.Printf(message)
			return -1, fmt.Errorf(message)
		} else {
			// If we get things back, the errors will be encoded within the JSON.
			if resp.StatusCode < 200 || resp.StatusCode > 299 {
				message := fmt.Sprintf("PostJSON: bad status from POST to %v [%v]\n", uri, resp.Status)
				log.Printf(message)
				return magic_index, fmt.Errorf(message)
			} else {
				// Parse the response. Everything comes from ReVal in our current formulation.
				var dat RevalResponse
				body, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(body, &dat)
				if err != nil {
					message := fmt.Sprintf("PostJSON: could not unmarshal response body")
					log.Printf(message)
					return magic_index, fmt.Errorf(message)
				}
			}
		}
		// Close the body at function exit.
		defer resp.Body.Close()
	}

	magic_index += 1
	return magic_index, nil
}
