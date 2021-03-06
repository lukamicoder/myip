package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var regex = regexp.MustCompile("(?m)[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}")
var verbose = false
var interactive = false

var urls = []string{
	"myipinfo.net",
	"myip.dnsomatic.com",
	"icanhazip.com",
	"checkip.dyndns.org",
	"www.myipnumber.com",
	"myexternalip.com",
	"www.ipchicken.com",
	"ipecho.net/plain",
	"bot.whatismyipaddress.com",
	"smart-ip.net/myip",
	"checkip.amazonaws.com",
	"www.checkip.org",
	"checkip.dns.he.net",
	"whatismyip.akamai.com",
	"ifconfig.me/ip",
	"api.ipify.org",
	"wgetip.com",
	"ident.me",
}

func main() {
	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			switch arg {
			case "-v", "--verbose":
				verbose = true
			case "-i", "--interactive":
				interactive = true
			case "-h", "--help":
				fmt.Print(`
Usage: myip [options]
where options include:
	--local, -l		return the local IP address(es)
	--verbose, -v		enable verbose output
	--interactive, -i	enable interactive mode
	--help, -h		print this help message
`)
				return
			case "-l", "--local":
				printLocalIP()
				return
			default:
				fmt.Print(`Invalid option: '` + arg + `'
`)
				return
			}
		}
	}

	printIP()
}

func printIP() {
	var currentIP net.IP

	if interactive {
		fmt.Println("Select a url from the list:")
		for i, url := range urls {
			fmt.Printf("	[%v] %v\n", i + 1, url)
		}
		fmt.Printf("	[%v] custom url\n", len(urls) + 1)

		var number int
		_, err := fmt.Scanln(&number)
		if err != nil || number <= 0 || number > (len(urls) + 1) {
			fmt.Println("Incorrect value has been entered.")
			return
		}

		var url string
		if number == len(urls) + 1 {
			fmt.Println("Enter url:")
			_, err := fmt.Scanln(&url)
			if err != nil {
				fmt.Println("Incorrect url has been entered.")
				return
			}
			if url == "" {
				return
			}

		} else {
			url = urls[number - 1]
		}

		currentIP = getExternalIPByURL(url)
	} else {
		rand.Seed(time.Now().UTC().UnixNano())
		currentIP = getExternalIP()
	}

	if currentIP == nil {
		fmt.Println("No external IP address found.")
		return
	}

	fmt.Println(currentIP.String())
}

func printLocalIP() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		if verbose {
			fmt.Printf("%v\n", err)
		}

		return
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Println(ipnet.IP.String())
			}
		}
	}
}

func getExternalIP() net.IP {
	var currentIP net.IP
	for _, i := range rand.Perm(len(urls)) {
		if verbose {
			fmt.Printf("Connecting to %v...\n", urls[i])
		}

		url := "http://" + urls[i]

		content, err := getResponse(url)
		if err != nil {
			if verbose {
				fmt.Printf("%v\n", err)
			}
			continue
		}

		ip := regex.FindString(content)

		currentIP = net.ParseIP(ip)

		if currentIP != nil {
			return currentIP
		}

		fmt.Println("No valid IP address could be parsed from response.")
	}

	return currentIP
}

func getExternalIPByURL(url string) net.IP {
	if verbose {
		fmt.Printf("Connecting to %v...\n", url)
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	content, err := getResponse(url)
	if err != nil {
		if verbose {
			fmt.Printf("%v\n", err)
		}
		return nil
	}

	ip := regex.FindString(content)

	currentIP := net.ParseIP(ip)

	if currentIP != nil {
		return currentIP
	}

	return nil
}

func getResponse(url string) (string, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: time.Duration(3 * time.Second),
	}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
