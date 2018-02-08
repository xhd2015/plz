package must

import (
	"encoding/json"
	"github.com/v2pro/plz/test"
	"github.com/v2pro/plz/test/testify/assert"
	"runtime"
	"reflect"
	"strings"
)

func JsonEqual(expected string, actual interface{}) {
	t := test.CurrentT()
	t.Helper()
	var expectedObj interface{}
	err := json.Unmarshal([]byte(expected), &expectedObj)
	if err != nil {
		t.Fatal("expected json is invalid: " + err.Error())
		return
	}
	var actualJson []byte
	switch actualVal := actual.(type) {
	case string:
		actualJson = []byte(actualVal)
	case []byte:
		actualJson = actualVal
	default:
		actualJson, err = json.Marshal(actual)
		t.Fatal("actual can not marshal to json: " + err.Error())
		return
	}
	var actualObj interface{}
	err = json.Unmarshal(actualJson, &actualObj)
	if err != nil {
		t.Log(string(actualJson))
		t.Fatal("actual json is invalid: " + err.Error())
		return
	}
	maskAnything(variables{}, expectedObj, actualObj)
	if assert.Equal(t, expectedObj, actualObj) {
		return
	}
	t.Helper()
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatal("check failed")
		return
	}
	t.Log(string(actualJson))
	t.Fatal(test.ExtractFailedLines(file, line))
}

type variable struct {
	value interface{}
	subs  []func(value interface{})
}

type variables map[string]*variable

func (vars variables) sub(varName string, sub func(value interface{})) {
	v := vars[varName]
	if v == nil {
		vars[varName] = &variable{subs: []func(value interface{}){sub}}
		return
	}
	if len(v.subs) == 0 {
		sub(v.value)
		return
	}
	v.subs = append(v.subs, sub)
}

func (vars variables) bind(varName string, varValue interface{}) {
	v := vars[varName]
	if v == nil {
		vars[varName] = &variable{value: varValue}
		return
	}
	if len(v.subs) > 0 {
		for _, sub := range v.subs {
			sub(varValue)
		}
		v.value = varValue
		v.subs = nil
		return
	}
	Equal(v.value, varValue)
}

func maskAnything(vars variables, expected interface{}, actual interface{}) {
	switch reflect.TypeOf(expected).Kind() {
	case reflect.Map:
		if reflect.ValueOf(actual).Kind() != reflect.Map {
			return
		}
		expectedVal := reflect.ValueOf(expected)
		actualVal := reflect.ValueOf(actual)
		keys := expectedVal.MapKeys()
		for _, keyIter := range keys {
			key := keyIter
			varName, _ := key.Interface().(string)
			if strings.HasPrefix(varName, "{") && strings.HasSuffix(varName, "}") {
				vars.sub(varName, func(value interface{}) {
					elem := expectedVal.MapIndex(key)
					expectedVal.SetMapIndex(key, reflect.ValueOf(nil))
					expectedVal.SetMapIndex(reflect.ValueOf(value), elem)
				})
			}
			actualElem := actualVal.MapIndex(key)
			if !actualElem.IsValid() {
				continue
			}
			expectedElem := expectedVal.MapIndex(key).Interface()
			varName, _ = expectedElem.(string)
			if strings.HasPrefix(varName, "{") && strings.HasSuffix(varName, "}") {
				expectedVal.SetMapIndex(key, actualElem)
				vars.bind(varName, actualElem.Interface())
				continue
			}
			maskAnything(vars, expectedElem, actualElem.Interface())
		}
	}
}
