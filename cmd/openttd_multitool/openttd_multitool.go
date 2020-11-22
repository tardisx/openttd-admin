package main

import (
	"flag"
	"github.com/tardisx/openttd-admin/pkg/admin"
	"os"
	"strings"
)

const currentVersion = "0.02"

type dailyFlags []string
type monthlyFlags []string
type yearlyFlags []string

func (i *dailyFlags) String() string {
	// change this, this is just can example to satisfy the interface
	return "my string representation"
}

func (i *dailyFlags) Set(value string) error {
	*i = append(*i, strings.TrimSpace(value))
	return nil
}

func (i *monthlyFlags) String() string {
	// change this, this is just can example to satisfy the interface
	return "my string representation"
}

func (i *monthlyFlags) Set(value string) error {
	*i = append(*i, strings.TrimSpace(value))
	return nil
}

func (i *yearlyFlags) String() string {
	// change this, this is just can example to satisfy the interface
	return "my string representation"
}

func (i *yearlyFlags) Set(value string) error {
	*i = append(*i, strings.TrimSpace(value))
	return nil
}

func main() {

	var daily dailyFlags
	var monthly monthlyFlags
	var yearly yearlyFlags

	flag.Var(&daily, "daily", "An RCON command to run daily - may be repeated")
	flag.Var(&monthly, "monthly", "An RCON command to run monthly - may be repeated")
	flag.Var(&yearly, "yearly", "An RCON command to run yearly - may be repeated")

	var hostname string
	var password string
	var port int
	flag.StringVar(&hostname, "hostname", "localhost", "The hostname (or IP address) of the OpenTTD server to connect to")
	flag.StringVar(&password, "password", "", "The password for the admin interface ('admin_password' in openttd.cfg)")
	flag.IntVar(&port, "port", 3977, "The port number of the admin interface (default is 3977)")
	flag.Parse()

	if password == "" {
		println("ERROR: You must supply a password")
		os.Exit(1)
	}

	server := admin.OpenTTDServer{}

	for _, value := range daily {
		server.RegisterDateChange("daily", value)
	}

	for _, value := range monthly {
		server.RegisterDateChange("monthly", value)
	}

	for _, value := range yearly {
		server.RegisterDateChange("yearly", value)
	}

	// this blocks forever
	server.Connect(hostname, port, password, "openttd-multitool", currentVersion)
}
