package main

import (
	"fmt"
	"reflect"
)

type A struct{ A string }

type B struct{ B string }

type Shape struct {
	Name   string
	Type   string
	Fields *Shape
}

func main() {
	description := Shape{Name: "description", Type: "string", Fields: nil}
	code := Shape{Name: "code", Type: "structure", Fields: &Shape{Name: "s3BucketName"}}

	fmt.Println(reflect.TypeOf(description))
	fmt.Println(reflect.TypeOf(code))
}

func main2() {
	a := A{}
	b := B{}

	fmt.Println(reflect.TypeOf(a))
	fmt.Println(reflect.TypeOf(b))
}
