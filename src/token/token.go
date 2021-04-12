package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

// jwt key
var mySigningKey = []byte("my_secret_key")

var myPort string

func HandleRequests() {
    router := mux.NewRouter()
    router.HandleFunc("/token", Token).Methods("GET")


	log.Fatal(http.ListenAndServe(":"+myPort, router))
}

func main() {
    myPort = "3333"

    HandleRequests()
}

// Token : curl -v -X GET localhost:3333/token
func Token(w http.ResponseWriter, r *http.Request) {
    validToken, err := GenerateJWT()
    if err != nil {
        fmt.Println("Failed to generate token")
    }
    // enclose token and return
	if enc, err := json.Marshal(validToken); err  == nil {
		w.WriteHeader( http.StatusOK )
		w.Write( []byte( enc ) )
	} else {
		w.WriteHeader( http.StatusInternalServerError )
	}
}

// GenerateJWT : generates a jwt token according to the agreen signing key.
func GenerateJWT() (string, error) {
    token := jwt.New(jwt.SigningMethodHS256)

    // create claims
    claims := token.Claims.(jwt.MapClaims)

    // set claims
    claims["authorized"] = true
    claims["client"] = "EasyRide"
    claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

    // create token string
    tokenString, err := token.SignedString(mySigningKey)

    if err != nil {
        fmt.Errorf("Something Went Wrong: %s", err.Error())
        return "", err
    }

    return tokenString, nil
}

