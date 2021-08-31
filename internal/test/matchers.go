package test

import (
	"reflect"

	"github.com/golang/mock/gomock"
)

type ofType struct{ t string }

func OfType(t string) gomock.Matcher {
	return &ofType{t}
}

func (o *ofType) Matches(x interface{}) bool {
	return reflect.TypeOf(x).String() == o.t
}

func (o *ofType) String() string {
	return "is of type " + o.t
}
