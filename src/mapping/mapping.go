package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/gorilla/mux"
	"googlemaps.github.io/maps"
)

var myPort string

// A Journey represents a journey that is built from the response to the google directions API
type Journey struct {
	Origin string
	Destination string
	DistanceTotal int
	DistanceA int
}

func main() {

	myPort = "1111"
	HandleRequests()

}


func HandleRequests() {
	router := mux.NewRouter()
	
	router.HandleFunc("/route", Route).Methods("POST")


	log.Fatal(http.ListenAndServe(":"+myPort, router))
}

// Route : recieves requests from ride microservice and retrieves route from google API.
func Route(w http.ResponseWriter, r *http.Request) {

	var journey Journey

	// decode request body into a journey struct.
	err := json.NewDecoder(r.Body).Decode(&journey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// create request
	c, err := maps.NewClient(maps.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		log.Fatalf("fatal error: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	// set params
	req := &maps.DirectionsRequest{
		Region:      "UK",
		Origin:      journey.Origin,
		Destination: journey.Destination,
	}
	// make request
	route, _, err := c.Directions(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		// handle data and return
		journey.DistanceTotal = route[0].Legs[0].Distance.Meters
		journey.DistanceA = calcARoadDistance(route)

		// if journey not successfully retrieved
		if journey.DistanceTotal == 0 {
			err2 := errors.New("invalid location")
			http.Error(w, err2.Error(), http.StatusBadRequest)
		} else {
			// enclose and write to response
			if enc, err := json.Marshal(journey); err == nil {
				w.WriteHeader(http.StatusOK)
				w.Write( []byte(enc) )

			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}
}

// calcARoadDistance : Calculate return the distance of A roads within a journey
func calcARoadDistance(s []maps.Route) int {
	distanceA := 0
	for _, step := range s[0].Legs[0].Steps {
		regexA, _ := regexp.Compile("A([0-9]+)")
		if regexA.MatchString(step.HTMLInstructions) == true {
			distanceA = distanceA + step.Distance.Meters
		}
	}
	return distanceA
}

