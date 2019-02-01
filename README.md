# Description
A simple command line bot for voting on strawpoll.me automatically written in Go.
By default uses 4242 HTTP proxies from the proxies.json file but it's easy to add more.

# Usage
```
Usage: main.go [flags]

Options:
  -poll
            The poll ID obtained from the URL.
            No default, must be included.
  -options
            Which option you would like to vote on.
            No default, must be included.
  -threads
            How many threads to be simultaneously sending votes on.
            Default: 100.
  -entrance
            Which proxy to start on (used if sending vote spam two parts).
            Default: 0.
  -timeout
            How long the HTTP client's timeout should be in seconds.
            Default: 30.
  -clean
            Whether we should clean the proxy list of dead proxies.
            Default: false.
```

### Example
```
Strawpoll-Bot.exe -poll 17338883 -options 139529712 -threads 100 -entrance 0 -timeout 30 -clean false
```

## How to obtain poll and options

### Poll
The poll ID is simply obtained from the end of the URL of the poll you want to bot.

From `https://www.strawpoll.me/17338883` the poll ID would be `17338883`.

### Options
The options are slightly harder to get: you need to use inspect element.
Find the vote you would like to bot, right click on it, then inspect element.
From here find the `<input>` tag associated with the vote you want.
Then find the `value="139529712"` inside the `<input>` tag.
Copy that value and that is the option.

I haven't yet coded in multiple options; I may do it in the future.
If you really want it now feel free to add it in yourself then make sure to do a pull request and I will approve it (after checking it of course).

# How to install

### Binary (easiest)
1. Download the binary obtained from [the releases page](../../releases).
2. Unzip the binary.
3. Open command prompt or terminal and cd to the location you downloaded the binary to.
4. Run a command built from the example and usage above. If you're on Linux, replace `Strawpoll-Bot.exe` with `./Strawpoll-Bot`.

### Source (if you have Go installed)
1. Download the source from the latest release from [the releases page](../../releases) or download the project.
2. Unzip the file (if zipped).
3. Open command prompt or terminal and cd to the location you downloaded the source to.
4. Run a command built from the example and usage above; replace `Strawpoll-Bot.exe` with `go run main.go`.

# Proxy list

### Clean
If you would like to remove all dead proxies from the file use the `-clean true` flag.
This will output a `clean-proxies.json` file once completed which you should use to replace the original `proxies.json` file.
Keep in mind this will decrease the speed of running the bot (quite significantly) but will increase the speed by not wasting time on dead links in the future.

### Update
If you would like to update the proxy list with your own simply replace the proxies replace the `proxies.json` file with your own JSON array of strings.
These strings must be the proxy in this format: `IP:Port`, for example: `127.0.0.1:1234`.
