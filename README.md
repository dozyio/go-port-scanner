# GO TCP Connect Port Scanner

A TCP connect port scanner written in Go. Uses concurrency to improve the speed.

[![asciicast](https://asciinema.org/a/FH9fRfWE0973Vv5o0r0tLqSfw.svg)](https://asciinema.org/a/FH9fRfWE0973Vv5o0r0tLqSfw?autoplay=1)

## Build
```sh
go build -o port-scanner
```

## Run
```sh
./port-scanner -s 1 -e 65535 -c 6 -w 1 192.168.1.253
```

## Usage

-s Start port number
-e End port number
-c Number of workers
-w Wait timeout

