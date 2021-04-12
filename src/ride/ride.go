package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var myPort string
var rosterPort string
var mappingPort string



// A Ride represents the Driver and Cost that a ride will take.
type Ride struct {
	Driver Driver
	Cost float32

}

// A Driver represents a driver, containing a name and a rate (pennies per km)
type Driver struct {
	Name string
	Rate int
}

// A Journey represents a journey that is extracted from the mapping microservice.
type Journey struct {
	Origin string
	Destination string
	DistanceTotal int
	DistanceA int
}

func main() {
	myPort = "4444"
	HandleRequests()
}

// HandleRequests handles HTTP requests
func HandleRequests() {
	router := mux.NewRouter()
	
	router.HandleFunc("/ride", FindRide).Methods("POST")

	log.Fatal(http.ListenAndServe(":"+myPort, router))
}


// requestDriver : requests a Driver from the roster, providing the cheapest driver on the roster at that time.
func requestDriver() (Driver, error) {
	url := os.Getenv("ROSTER_SERVICE_URL") + "/drivers/cheapest"
	client := &http.Client {}
	var empty Driver
	if req, err := http.NewRequest( "GET", url, nil ); err == nil {
		if resp, err1 := client.Do( req );
			err1 == nil {
			if cheapestDriver, err2 := ioutil.ReadAll( resp.Body ); 
				err2 == nil	{
				var driver Driver
				json.Unmarshal(cheapestDriver, &driver)

				// if driver successfully found
				if driver.Name != "" {
					return driver, nil
				}
				
			} else {
				return empty, err2
			}
		} else {
			return empty, err1
		}
	} else {
		return empty, err
	}
	return empty, errors.New("roster empty")
}



// requestJourney : takes a riders origin and destination, and requests a journey from the mapping serivce.
func requestJourney(origin string, destination string) (Journey, error)  {

	// url := "http://localhost:"+ mappingPort + "/route"
	url := os.Getenv("MAPPING_SERVICE_URL") + "/route"
	client := &http.Client {}
	data := map[ string ] string{ "Origin" : origin, "Destination" : destination }

	var empty Journey

	// enclose data to be sent as json object
	if enc, err := json.Marshal( data ); err == nil {
		// create request
		if req, err1 := http.NewRequest( "POST", url, bytes.NewBuffer( enc ) );
			err1 == nil {
			// make request
			if resp, err2 := client.Do( req );
				err2 == nil {
				if route, err3 := ioutil.ReadAll( resp.Body );
					err3 == nil {
					
					var journey Journey
					// unmarshall response into journey struct
					json.Unmarshal(route, &journey)
					
					// if journey successfully obtained
					if journey.DistanceTotal != 0 {
						return journey, nil
						
					} else {
						err4 := errors.New("invalid location")
						return empty, err4
					}

					} else {
						return empty, err3
					}
				} else {
					return empty, err2
				}
			} else {
				return empty, err1
			}
	} else {
		return empty, err
	}	
}

// request RosterSize retrieves the size of the roster from the roster microservice.
func requestRosterSize() (int, error) {
	url := os.Getenv("ROSTER_SERVICE_URL") + "/rostersize"
	client := &http.Client {}

	// create request
	if req, err := http.NewRequest( "GET", url, nil );
		err == nil {
		// make request
		if resp, err1 := client.Do( req );
			err1 == nil {
			if rosterSize, err2 := ioutil.ReadAll( resp.Body ); 
				err2 == nil	{
				// convert to int
				rosterSize, _ := strconv.Atoi(string(rosterSize))
				return rosterSize, nil
				
			} else {
				return 0, err2
			}
		} else {
			return 0, err1
		}
	} else {
		return 0, err
	}
}

// calculatePrice applies the surge pricing algorithm to the journey, calculating the final cost of a ride
func calculatePrice(distanceA int, distanceTotal int, rate int, numDrivers int) float32 {
	// initial multiplier set to 0
	multiplier := 0

	// Check if the majority of the journey is made on A roads
	if distanceA > distanceTotal/2 {
		multiplier += 2
	}
	//  +2 if the journey is to begin between 23:00 and 06:00.
	start := time.Now()
	if start.Hour() < 6 || start.Hour() == 23  {
		multiplier += 2
	}
	// Check if number of drivers on roster is less than 5
	if numDrivers < 5 {
		multiplier += 2
	}

	// return cost
	if multiplier == 0 {
		return float32(rate*(distanceTotal/1000))/100
	} else {
		return float32(rate*(distanceTotal/1000)*multiplier)/100
	}
	

}


// FindRide : curl -X POST localhost:4444/ride --data '{"Origin" : "Exeter", "Destination" : "Plymouth"}'
func FindRide(w http.ResponseWriter, r *http.Request) {

	var ride Ride
	var journey Journey
	// Try to decode the request body into the struct
	err := json.NewDecoder(r.Body).Decode(&journey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// if Driver successfully obtained
	if driver, err := requestDriver(); err == nil {
		// If journey successfully obtained
		if journey, err1 := requestJourney(journey.Origin, journey.Destination); err1 == nil {
			// if roster size successfully obtained
			if size, err2 := requestRosterSize(); err2 == nil {
				ride.Cost = calculatePrice(journey.DistanceA, journey.DistanceTotal, driver.Rate, size)
				ride.Driver = driver

				// enclose ride into json response
				if enc, err := json.Marshal(ride); err  == nil {
					w.WriteHeader( http.StatusOK )
					w.Write( []byte( enc ) )
				} else {
					w.WriteHeader( http.StatusInternalServerError )
					}
			} else { 
				w.WriteHeader(http.StatusNotFound)
				http.Error(w, err2.Error(), 404)
				fmt.Printf( "GET to Roster failed with %s\n", err2 )
			}
		} else {
			http.Error(w, err1.Error(), http.StatusBadRequest)
			fmt.Printf( "POST to Mapping failed with %s\n", err1 )
		} 

	} else {
		w.WriteHeader(http.StatusNotFound)
		http.Error(w, err.Error(), 404)
		fmt.Printf( "GET to Roster failed with %s\n", err )
	}




}