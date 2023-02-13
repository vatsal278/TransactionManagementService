package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	// Define the URL of the service you want to call
	url := "http://localhost:80/microbank/v1/user"

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
	}
	c := http.Cookie{
		Name:  "token",
		Value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZGJkZmZmZmYtNmQ0Yi00YTAzLTg5YWItOTkxNDM4ZDNkZGY4IiwiZXhwIjoxNjc0NjQ4NDI4LCJpYXQiOjE2NzQ2NDgxMjh9.Mwxm1AaJQ6QMe4kKpJtkCZ3FC0W2eMayUeuHXoL5iqA",
	}
	req.AddCookie(&c)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	// Print the response body
	fmt.Println(string(body))
}
