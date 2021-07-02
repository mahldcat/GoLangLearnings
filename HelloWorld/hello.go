package main

import (
	"fmt"
	"log"

	"example.com/mymodule"
	"rsc.io/quote"
)

func HandleCall(input string) {
	msg, err := mymodule.Hello(input)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(msg)
	}

}

func main() {
	log.SetPrefix("mymodule: ")
	log.SetFlags(0)

	names := []string{"Foo", "Bar", "Zam", "Zok"}

	msgs, err := mymodule.MultiHello(names)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(msgs)

	//HandleCall("foobletch")
	//HandleCall("")

	fmt.Println(quote.Go())
}
