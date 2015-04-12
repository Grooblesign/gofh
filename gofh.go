package main

import "fmt"
// import "encoding/json"

type Person struct {
    surname string
    forenames string
    gender string
	fatherId int64
	motherId int64
}

func main() {
    fmt.Printf("hello, world\n")
	
	paul := Person{"Garner", "Paul", "Male", 0, 0}
	
	paul.forenames = "Paul John"
	
	fmt.Printf(paul.forenames + " " + paul.surname)
	
	// b, err := json.Marshal(paul)
	
	// fmt.Printf(b)
}