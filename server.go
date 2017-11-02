package main

import (
	"fmt"
	"os"

	"github.com/fragmenta/server"
	"github.com/fragmenta/server/config"

	"github.com/fragmenta/fragmenta-cms/src/app"
)

// Main entrypoint for the server which performs bootstrap, setup
// then runs the server. Most setup is delegated to the src/app pkg.
func main() {

	// Bootstrap if required (no config file found).
	if app.RequiresBootStrap() {
		err := app.Bootstrap()
		if err != nil {
			fmt.Printf("Error bootstrapping server %s\n", err)
			return
		}
	}

	// Setup our server
	server, err := SetupServer()
	if err != nil {
		fmt.Printf("server: error setting up %s\n", err)
		return
	}

	// If in production on port 443, set up a server instead
	if server.Production() && server.Port() == 443 {

		// Redirect http traffic to https
		server.StartRedirectAll(80, server.Config("root_url"))

		// Serve https directly using autocert
		err = server.StartTLSAutocert(server.Config("autocert_email"), server.Config("autocert_domains"))
		if err != nil {
			server.Fatalf("Error starting server %s", err)
		}

	} else {
		// Start the server using http
		err = server.Start()
		if err != nil {
			server.Fatalf("Error starting server %s", err)
		}
	}

}

// SetupServer creates a new server, and delegates setup to the app pkg.
func SetupServer() (*server.Server, error) {

	// Setup server
	s, err := server.New()
	if err != nil {
		return nil, err
	}

	// Load the appropriate config
	c := config.New()
	err = c.Load("secrets/fragmenta.json")
	if err != nil {
		return nil, err
	}
	config.Current = c

	// Check environment variable to see if we are in production mode
	if os.Getenv("FRAG_ENV") == "production" {
		config.Current.Mode = config.ModeProduction
	}

	// Call the app to perform additional setup like mail, assets, views, auth and routes
	app.Setup()

	return s, nil
}
