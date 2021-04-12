package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// A Driver represents a driver that is kept within the roster
type Driver struct {
	Name string
	Rate int
}

// jwt key
var mySigningKey = []byte("my_secret_key")

var myPort string

// Drivers is a map that contains all drivers on the roster
var Drivers = make(map[string] Driver)



func main() {
	myPort = "2222"
	populate()
	HandleRequests()
}

// populate roster
func populate() {
	maximilian := Driver{
		Name: "Max",
		Rate: 20,
	}
	julie := Driver{
		Name: "Julie",
		Rate: 40,
	}
	ron := Driver{
		Name: "Ron",
		Rate: 35,
	
	}

	// Create IDs and add to roster
	if id, err := uuid.NewUUID(); err == nil {
		Drivers[id.String()] = maximilian
	}

	if id, err := uuid.NewUUID(); err == nil {
		Drivers[id.String()] = julie
	}

	if id, err := uuid.NewUUID(); err == nil {
		Drivers[id.String()] = ron
	}
}

func HandleRequests() {
	router := mux.NewRouter()
	
	// GET requests, requiring no authorization
	router.HandleFunc("/rostersize", RosterSize).Methods("GET")
	router.HandleFunc("/drivers", ListDrivers).Methods("GET")
	router.HandleFunc("/drivers/cheapest", CheapestDriver).Methods("GET")
	router.HandleFunc("/drivers/{id}", ListDriver).Methods("GET")

	// requests requiring authorization
	router.Handle("/drivers", isAuthorized(NewDriver)).Methods("POST")
	router.Handle("/drivers/{id}", isAuthorized(UpdateDriver)).Methods("PUT")
	router.Handle("/drivers/{id}", isAuthorized(DeleteDriver)).Methods("DELETE")


	log.Fatal(http.ListenAndServe(":"+myPort, router))
}


// isAuthorized checks that jwt Token is present, and verifies that it is valid.
func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		
        if r.Header["Token"] != nil {

			// check token is valid
            token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, fmt.Errorf("There was an error")
                }
                return mySigningKey, nil
            })

            if err != nil {
                fmt.Fprintf(w, err.Error())
            }

            if token.Valid {
				fmt.Println("Authorized")
                endpoint(w, r)

            }
        } else {

            fmt.Fprintf(w, "Not Authorized")
        }
    })
}

// RosterSize : curl -v -X GET localhost:2222/rostersize 
func RosterSize(w http.ResponseWriter, r *http.Request) {

	// package roster size and return request
	if enc, err := json.Marshal( len(Drivers) ); err == nil {
		w.WriteHeader( http.StatusOK)
		w.Write( []byte(enc))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}	
}

// ListDrivers :  curl -v -X GET localhost:2222/drivers
func ListDrivers(w http.ResponseWriter, r *http.Request) {

	// enclose roster to json and return
	if enc, err := json.Marshal( Drivers ); err == nil {
		w.WriteHeader( http.StatusOK)
		w.Write( []byte(enc))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ListDriver : curl -v -X GET localhost:2222/drivers/{id}
func ListDriver(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// enclose driver to json and return
	if enc, err := json.Marshal(Drivers[id]); err == nil {
		w.WriteHeader(http.StatusOK)
		w.Write( []byte(enc) )
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// NewDriver : curl -v -X POST localhost:2222/drivers --header 'Token:xxx' --data '{"Name":"...",Rate:...}'
func NewDriver(w http.ResponseWriter, r *http.Request) {

	// create id
	if id, err := uuid.NewUUID(); err == nil {
		decoder := json.NewDecoder(r.Body)
		var driver Driver
		// add new driver
		if err := decoder.Decode( &driver ); err == nil {
			w.Header().Set( "Location", r.Host + "/Drivers/" + id.String())
			w.WriteHeader(http.StatusCreated)
			Drivers[id.String()] = driver

			// enclose new driver id to json and return
			if enc, err := json.Marshal(id); err == nil {
				w.WriteHeader(http.StatusOK)
				w.Write( []byte(enc) )
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}

		} else {
			// handle invalid input
			w.WriteHeader(http.StatusBadRequest)
			err := errors.New("invalid driver data")
			http.Error(w, err.Error(), 404)

		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// UpdateDriver : curl -v -X PUT localhost:2222/drivers/{id} --header 'Token:xxx' --data '{"Name":"...",Rate:...}'
func UpdateDriver(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// check driver exists
	if _, ok := Drivers[id]; ok {
		decoder := json.NewDecoder( r.Body )
		var driver Driver

		// decode input into a driver struct
		if err := decoder.Decode( &driver ); err == nil {
			// update driver in roster with new data
			Drivers[id] = driver
		} else{
			w.WriteHeader(http.StatusBadRequest)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}
// DeleteDriver : curl -v -X DELETE localhost:2222/drivers/{id} --header 'Token:xxx'
func DeleteDriver(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// check driver exists, then delete
	if _, ok := Drivers[id]; ok {
		w.WriteHeader(http.StatusNoContent)
		delete(Drivers, id)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// CheapestDriver : curl -v -X GET localhost:2222/drivers/cheapest
func CheapestDriver(w http.ResponseWriter, r *http.Request)  {
	
	// if roster not empty
	if len(Drivers) > 0 {
		var cheapestDriver Driver
		cheapestDriver.Rate = 9999999999999

		// find cheapest driver
		for _, driver := range Drivers {
			if driver.Rate < cheapestDriver.Rate {
				cheapestDriver = driver
			}
		}
		// enclose cheapest driver within json and return
		if enc, err := json.Marshal(cheapestDriver); err == nil {
			w.WriteHeader(http.StatusOK)
			w.Write( []byte(enc) )
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		// roster empty
		w.WriteHeader(http.StatusBadRequest)
	}

}
