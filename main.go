package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var (
	// ProxyList is the array of proxies formatted as: "IP:Port".
	ProxyList []string
	// Poll is the poll's ID, obtained from the URL with this format: "https://www.strawpoll.me/ID".
	Poll string
	// Options are the options used in the poll.
	Options string
	// Threads is how many post threads will be running.
	Threads int
	// Entrance is which proxy to start from.
	Entrance int
	// CleanProxies determines whether we should cleanse the proxy list of dead proxies.
	CleanProxies bool
	// SecondTimeout is the HTTP client timeout duration in seconds.
	SecondTimeout int
	// CleanChannel is the channel where working proxies are put (if CleanProxies is true).
	CleanChannel = make(chan *url.URL)
	// ProxyChannel is the channel for proxies.
	ProxyChannel = make(chan *url.URL)
	// QuitChannel is the channel used to kill voting threads.
	QuitChannel = make(chan bool)
	// QuitCleanChannel is the channel used to kill (surprisingly enough) the clean thread.
	QuitCleanChannel = make(chan bool)
	// CompletedChannel is the channel used by the main thread to signal all channels are dead.
	CompletedChannel = make(chan bool)
)

// ReadFlags reads the flags into global variables.
func ReadFlags() {
	// Just for debugging.
	log.Println("Reading flags!")

	// Botting the example poll looked like this: "Strawpoll-Bot.exe -poll 17338883 -options 139529712 -threads 100 -entrance 0 -timeout 30 -clean true"
	flag.StringVar(&Poll, "poll", "", "The poll ID obtained from the URL.")                                 // Example: "-poll 17338883".
	flag.StringVar(&Options, "options", "", "Which option you would like to vote on.")                      // Example: "-options 139529712".
	flag.IntVar(&Threads, "threads", 100, "How many threads to be simultaneously sending votes on.")        // Example: "-threads 100".
	flag.IntVar(&Entrance, "entrance", 0, "Which proxy to start on (used if sending vote spam two parts).") // Example: "-entrance 312".
	flag.IntVar(&SecondTimeout, "timeout", 30, "How long the HTTP client's timeout should be in seconds.")  // Example: "-timeout 30".
	flag.BoolVar(&CleanProxies, "clean", false, "Whether we should clean the proxy list of dead proxies.")  // Example: "-clean false".
	flag.Parse()
}

// ReadProxyList reads the proxy list file to the ProxyList string array and returns any errors.
func ReadProxyList() (err error) {
	// Just for debugging.
	log.Println("Reading proxy list!")

	// Open the JSON file.
	proxyListFile, err := os.Open("proxies.json")
	if err != nil {
		return
	}

	// Parse the JSON data into an array.
	err = json.NewDecoder(proxyListFile).Decode(&ProxyList)
	return
}

// CleanThread is the thread for saving clean proxies.
func CleanThread() {
	var proxies []string

	// Run this loop until return or break is called.
	for {
		// Read from whichever comes first: a quit message or a proxy.
		select {
		case <-QuitCleanChannel:
			// We got the quit message; save clean proxies then exit.
			bytes, err := json.Marshal(proxies)
			if err != nil {
				// There was an error marshalling the JSON; log it and return.
				log.Printf("Error marshalling JSON: %v", err)
				CompletedChannel <- true
				return
			}

			err = ioutil.WriteFile("clean-proxies.json", bytes, 644)
			if err != nil {
				// There was an error writing the file; log it and return.
				log.Printf("Error writing file: %v", err)
			}

			CompletedChannel <- true
			return

		case proxy := <-CleanChannel:
			// We got a clean proxy; save it into an array.
			proxies = append(proxies, proxy.Host)
		}
	}
}

// PostThread is a thread for posting results.
func PostThread() {
	// Run this loop until return or break is called.
	for {
		// Read from whichever comes first a quit message, or a proxy.
		select {
		case <-QuitChannel:
			// There are no more proxies to use, this thread should now die.
			return

		case proxy := <-ProxyChannel:
			// Create the HTTP client with the transport proxy requested from the proxy channel.
			client := &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxy),
				},

				Timeout: time.Duration(SecondTimeout) * time.Second,
			}

			// Create a form with the options.
			form := url.Values{}
			form.Add("options", Options)

			// Send POST request with the client (using the requested proxy) and the form.
			_, err := client.PostForm("https://www.strawpoll.me/"+Poll, form)
			if err != nil {
				// If there is an error: log it and continue the loop.
				log.Printf("Error posting form: %v", err)
				continue // This name is kind of misleading, it will end this loop here but not exit out of the for loop.
			}

			// If we should clean the proxies:
			if CleanProxies {
				// Put the proxy into the clean channel as there have been no errors with it!
				CleanChannel <- proxy
			}
		}
	}
}

func main() {
	ReadFlags()

	// Read the proxy list into the ProxyList array.
	err := ReadProxyList()
	if err != nil {
		log.Printf("Error reading proxy list: %v", err)
		return
	}

	// If we are cleaning the proxies: start a clean thread!
	if CleanProxies {
		log.Println("Creating clean proxies thread!")
		go CleanThread()
	}

	// Create a post thread for how many threads requested in the flags.
	for i := 0; i < Threads; i++ {
		log.Printf("Creating thread %v!", i+1)
		go PostThread()
	}

	// Just for debugging.
	log.Println("All threads successfully created!")
	log.Println("Starting to read proxies / send votes!")

	// Run range for every proxy in the list (starting at entrance position).
	for i, proxy := range ProxyList[Entrance:] {
		// Parse the URL with format "http://proxy", example: "http://127.0.0.1:1234"
		proxyURL, err := url.Parse("http://" + proxy)
		if err != nil {
			// If there is an error: log it and continue the loop.
			log.Printf("Error parsing proxy URL: %v", err)
			continue // This name is kind of misleading: it will end this loop here, but not exit out of the for loop.
		}

		// Place proxy into ProxyChannel to be read by one of the threads created above.
		ProxyChannel <- proxyURL

		// Log which proxy we are on in the array.
		/* The reason it is i+1+Entrance is because:
		   i starts at 0 but it should debug as "Sending vote 1!" first.
		   Entrance will offset where to start in the array so that should be added on too. */
		log.Printf("Sending vote %v!", i+1+Entrance)
	}

	// Just for debugging.
	log.Println("All proxies have been sent to the proxy channel!")
	log.Println("Killing threads!")

	// For each thread that exists, send the die message in the quit channel.
	for i := 0; i < Threads; i++ {
		log.Printf("Killing thread %v!", i+1)
		QuitChannel <- true
	}

	// All other threads have finished, kill the clean channel!
	if CleanProxies {
		log.Println("Killing clean channel!")
		QuitCleanChannel <- true
		<-CompletedChannel
	}
}
