# TCP Relay Server

The goal is to build a TCP Relay Server which can:
* Listen for apps wanting to be relayed
* When such an app requests a connection, respond with a newly created socket
* Listen for one or many clients, which can have many clients themselves
* Mediate that traffic
* Not initiate any connections itself

## Notes

It's currently assumed that the App will take care of handling Application
Layer Protocols such as HTTP/HTTPS/SSH from a byte stream.
Errors in critical places lead to program exit.

## Basic Flow

1. Relay is set up, will open App Registration port 8080.
2. Echo server will connect to 8080.
3. Relay app will open an externally open Relayed port 8081.
4. When a client connects to 8081, Relay will inform the App of it, and will ask for a new TCP connection to a new port from the App.
5. App will dial in on 8082.
6. Relay will transfer everything from the client to the app and back.
7. Every new client will make the App form a new TCP connection to the relay (steps 4-6).

## Example Usage

Run the commands in `example.sh` different terminals for better clarity.

## Adding Custom Relayed Apps

Current implementation gives a *bufio.ReadWriter for every Client connected
to the Relay server socket created for a given registered app.

Edit the `func customFunction(rw *bufio.ReadWriter){}` in `sampleappecho/echo.go` to
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

    * e.g. clients with multiple clients of their own

* Remove excessive logging
* Code review
* Integrate with CI/CD

* Further nail down requirements
* Based on requirements, prioritize new performance improvements / new features / usability etc.

* Build an example app handling HTTP requests
* Handle closing of connections more gracefully
