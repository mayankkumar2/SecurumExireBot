package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")

	http.HandleFunc("/bot", func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		fmt.Println(string(b))

		_, _ = w.Write([]byte("Hi"))
	})

	fmt.Println("listening at port... " + port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("error: " + err.Error())
	}
}