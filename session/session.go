package session

import (
	"fmt"
	"os"
	"io/ioutil"
	"encoding/json"
)

type Session struct {
	File           string
	Movie_id       int `json:"movie_id"`
}

func (this *Session) IsSessionExist() bool {	
	if _, err := os.Stat(this.File); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}


func (this *Session) ReadSession() {
	dat, err := ioutil.ReadFile(this.File)
	if err != nil {
		fmt.Println(err)
	}

	ok := json.Unmarshal(dat, this)

	if ok != nil {
		fmt.Println("error:", ok)
	}
}

func (this *Session) Remove() {
	err := os.Remove(this.File)
	if err != nil {
		fmt.Println(err)
	}
}


func (this *Session) CreateSession() {
	bytes, err := json.Marshal(this)
	if err != nil {
		fmt.Println("error:", err)
	}

	if this.IsSessionExist() {
		err := ioutil.WriteFile(this.File, bytes, 0644)
		if err != nil {
			fmt.Println(err)
		}
	} else {		
		file, err := os.Create(this.File)
		if err != nil {
			fmt.Println(err)
		}
		_, ok := file.Write(bytes)

		if ok != nil {
			fmt.Println(ok)
		}
	}
}

