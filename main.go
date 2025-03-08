package main

import (
	"fmt"
	"log"
)

func main() {
	c, err := NewMaxBytesCache[int](100000)
	if err != nil {
		log.Fatal(err)
	}

	i := 0
	for {
		fmt.Printf("[%d], dataSize: %d\n", i, c.dataSize())
		c.Set(fmt.Sprintf("%d", i), i)
		i++
	}
}
