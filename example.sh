#!/usr/bin/bash

# relay server setup
# relaystyle 0 - client side port will be random
# relaystyle 1 - client side port will be regport + n registerd apps + 1
go run main.go -regport 9999 -relaystyle 1 &
sleep 1

# register apps
go run main.go -sampleapp -regportecho 9999 &
sleep 1
go run main.go -sampleapp -regportecho 9999 &

## Run in terminal 1
# while true; do echo term1 ;  done | nc -q 1 -w 1 localhost 10000

## Run in terminal 2
# while true; do echo term2 ;  done | nc -q 1 -w 1 localhost 10000

## connect to relayedport by running:
# telnet localhost 10000