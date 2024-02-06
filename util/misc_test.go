package util

import "testing"

func TestPrint(t *testing.T) {
	Print("hello")
	Print(1)

	type AnotherPerson struct {
		Name string
	}

	type Person struct {
		Name          string `json:"badname"`
		AnotherPerson AnotherPerson
	}
	Print(Person{Name: "test", AnotherPerson: AnotherPerson{Name: "test-father"}})
}
