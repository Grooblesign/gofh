package main

import "fmt"
// import "encoding/json"

type Person struct {
	id: int64
    surname string
    forenames string
    gender string
	fatherId int64
	motherId int64
}

func main() {
    fmt.Printf("hello, world\n")
	
	john := Person{1, "Garner", "Frederick John", "Male", 0, 0}
	paul := Person{2, "Garner", "Paul", "Male", 1, 0}
	
	paul.forenames = "Paul John"
	
	fmt.Printf(paul.forenames + " " + paul.surname)
	
	// b, err := json.Marshal(paul)
	
	// fmt.Printf(b)
}