# Wordfeud

A Go library for interacting with the [Wordfeud](https://wordfeud.com) API.

## Installation
`go get -u github.com/kayex/wordfeud`

## Usage
```go
package main

import (
	"log"
	"github.com/kayex/wordfeud"
)

func main() {
	client := wordfeud.NewClient()
	session, err := client.LoginWithEmail("user@example.com", "password")
	if err != nil {
		log.Fatal(err)
	}

	games, err := client.Games(session)
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.Move(session, games[0].ID, []wordfeud.Placement{
		wordfeud.Place(7, 5, "H", false),
		wordfeud.Place(7, 6, "E", false),
		wordfeud.Place(7, 7, "L", false),
		wordfeud.Place(7, 8, "L", false),
		wordfeud.Place(7, 9, "O", false),
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.SendChatMessage(session, games[0].ID, "Well played!")
	if err != nil {
		log.Fatal(err)
	}
}
```

## License
MIT