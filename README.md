# TCP Relay Server

The goal is to build a TCP Relay Server which can:
* Listen for apps wanting to be relayed
* When such an app requests a connection, respond with a newly created socket
* Listen for one or many clients, which can have many clients themselves
* Mediate that traffic

## Example Usage

Run the commands in different terminals for better clarity.

```bash
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
```

## Adding Custom Relayed Apps

Current implementation gives a *bufio.ReadWriter for every Client connected
to the Relay server socket created for a given registered app.

Edit the `func customFunction(rw *bufio.ReadWriter){}` in `echo/echo.go` to
get the hang of what's passed to the App and how to customize responses.


## Testing

```bash
go test
```

Also,

```bash
go build -race
```

## Static Analysis

```bash
circleci-lint run
```

## Further development

* Add tests
* Remove excessive logging
* Code review
* Integrate with CI/CD

* Further nail down requirements
* Based on requirements, prioritize new performance improvements / new features / usability etc.
