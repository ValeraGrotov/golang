package main

import (
	"fmt"
	"session"
	"strconv"
)

func main() {
	var chat_id int64 = 111111111

	ses := session.Session{
		File: "sessions/" + strconv.FormatInt(chat_id, 10) + ".json",
		Movie_id: 500,
	}

	if ses.IsSessionExist() {

		ses.ReadSession()

		fmt.Println(ses.Movie_id)
	} else {
		ses.CreateSession()
	}
}
