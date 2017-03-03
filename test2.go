package main

import "fmt"

func main() {
    var x [10]int
    xLen := 0

    x[xLen] = 3
    xLen++

    x[xLen] = 7
    xLen++

    fmt.Printf("%d\n", len(x[0:xLen]))
}
