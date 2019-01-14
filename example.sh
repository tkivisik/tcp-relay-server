#!/usr/bin/bash

# relay server setup
go run main.go -regport 9999 &
sleep 1

# register apps
go run echo/echo.go -regport 9999 &
sleep 1
go run echo/echo.go -regport 9999 &

# connect to relayedport by running:
# telnet localhost <port shown to screen>
