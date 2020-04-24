package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
        "strconv"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
	flag "github.com/spf13/pflag"
)

// Claims stores the values we want to extract from the JWT as JSON
type Claims struct {
	Email string `json:"email"`
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var (
	// default flag values
	authDomain = getEnv("AUTH_DOMAIN", "")
	address    = getEnv("LISTEN_ADDRESS", "")
	port, err  = strconv.Atoi(getEnv("LISTEN_PORT", "80"))

	// jwt signing keys
	keySet oidc.KeySet
)

func init() {

	// parse flags
	flag.StringVar(&authDomain, "auth-domain", authDomain, "authentication domain (https://foo.cloudflareaccess.com)")
	flag.IntVar(&port, "port", port, "http port to listen on (default 80)")
	flag.StringVar(&address, "address", address, "http address to listen on (leave empty to listen on all interfaces)")
	flag.Parse()

        if port <= 0 {
                fmt.Printf("ERROR: Invalid port number %s \n", port)
                flag.Usage()
                os.Exit(1)
        }

	// --auth-domain is required
	if authDomain == "" {
		fmt.Println("ERROR: Please set --auth-domain to the authorization domain you configured on cloudflare. Should be like `https://foo.cloudflareaccess.com`")
		flag.Usage()
		os.Exit(1)
	}

	// configure keyset
	certsURL := fmt.Sprintf("%s/cdn-cgi/access/certs", authDomain)
	keySet = oidc.NewRemoteKeySet(context.TODO(), certsURL)

}

func main() {

	// set up routes
	router := httprouter.New()
	router.GET("/auth/:audience", authHandler)

	// listen
	addr := fmt.Sprintf("%s:%d", address, port)
	log.Printf("Listening on %s", addr)
	log.Fatalln(http.ListenAndServe(addr, handlers.LoggingHandler(os.Stdout, router)))

}

func authHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	// Get audience from request params
	audience := ps.ByName("audience")

	// Configure verifier
	config := &oidc.Config{
		ClientID: audience,
	}
	verifier := oidc.NewVerifier(authDomain, keySet, config)

	// Make sure that the incoming request has our token header
	//  Could also look in the cookies for CF_AUTHORIZATION
	accessJWT := r.Header.Get("Cf-Access-Jwt-Assertion")
	if accessJWT == "" {
		write(w, http.StatusUnauthorized, "No token on the request")
		return
	}

	// Verify the access token
	ctx := r.Context()
	idToken, err := verifier.Verify(ctx, accessJWT)
	if err != nil {
		write(w, http.StatusUnauthorized, fmt.Sprintf("Invalid token: %s", err.Error()))
		return
	}

	// parse the claims
	claims := Claims{}
	err = idToken.Claims(&claims)
	if err != nil {
		write(w, http.StatusUnauthorized, fmt.Sprintf("Invalid claims: %s", err.Error()))
		return
	}

	// email is required in claims
	if claims.Email == "" {
		write(w, http.StatusUnauthorized, "No email in JWT claims")
		return
	}

	// Request is good to go
	w.Header().Set("X-Auth-User", claims.Email)
	write(w, http.StatusOK, "OK!")

}

func write(w http.ResponseWriter, status int, body string) {
	w.WriteHeader(status)
	_, err := w.Write([]byte(body))
	if err != nil {
		log.Printf("Error writing body: %s\n", err)
	}
}
