package main

import "fmt"

type subStruct struct {
	testy bool
}

type testStruct struct {
	testing string
	sub     *subStruct
}

func main() {
	structs := []testStruct{
		{
			testing: "hello1",
			sub:     nil,
		},
		{
			testing: "hello2",
			sub: &subStruct{
				testy: true,
			},
		},
		{
			testing: "hello3",
		},
	}

	for _, tst := range structs {
		if tst.sub != nil && tst.sub.testy {
			fmt.Println(tst.testing)
		} else {
			fmt.Println("fail")
		}
	}
}
