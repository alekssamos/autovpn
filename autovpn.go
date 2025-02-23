package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"runtime"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	chosenCountry := "US"
	if len(os.Args) > 1 && len(os.Args[1]) == 2 {
		chosenCountry = os.Args[1]
	}
	var tmp string = ""
	URL := "http://www.vpngate.net/api/iphone/"

	fmt.Printf("[autovpn] getting server list\n")
	response, err := http.Get(URL)
	check(err)

	defer response.Body.Close()
	scanner := bufio.NewScanner(response.Body)

	fmt.Printf("[autovpn] parsing response\n")
	fmt.Printf("[autovpn] looking for %s\n", chosenCountry)

	counter := 0
	for scanner.Scan() {
		if counter <= 1 {
			counter++
			continue
		}
		splits := strings.Split(scanner.Text(), ",")
		if len(splits) < 15 {
			break
		}

		country := splits[6]
		conf, err := base64.StdEncoding.DecodeString(splits[14])
		if err != nil || chosenCountry != country {
			continue
		}

		fmt.Printf("[autovpn] writing config file\n")
		if runtime.GOOS == "windows" {
			tmp = os.Getenv("temp")+"\\"
			err = ioutil.WriteFile(tmp+"openvpnconf.ovpn", conf, 0)
		} else {
			tmp = "/tmp/"
			err = ioutil.WriteFile("openvpnconf.ovpn", conf, 0664)
		}
		check(err)
		fmt.Printf("[autovpn] running openvpn\n")


		var cmd = exec.Command("")
		if runtime.GOOS == "windows" {
			cmd = exec.Command("openvpn", "--connect", tmp+"openvpnconf.ovpn")
		} else {
			cmd = exec.Command("sudo", "openvpn", tmp+"openvpnconf.ovpn")
		}
		cmd.Stdout = os.Stdout

		c := make(chan os.Signal, 2)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			cmd.Process.Kill()
		}()

		cmd.Start()
		cmd.Wait()

		fmt.Printf("[autovpn] try another VPN? (y/n) ")
		var input string
		fmt.Scanln(&input)
		if strings.ToLower(input) == "n" {
			os.Exit(0)
		}
	}
}
