package structmap_test

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/amulets/structmap"
	"github.com/amulets/structmap/behavior"
	"github.com/amulets/structmap/behavior/cast"
	"github.com/amulets/structmap/behavior/flag"
	"github.com/amulets/structmap/behavior/name"
	"github.com/amulets/structmap/internal"
)

type SubSubStruct struct {
	Address string
	Number  *int
}

type SubStruct struct {
	*SubSubStruct
	Age *string `structmap:",required" default:"15"`
}

type MyStruct struct {
	SubStruct
	MyAddress SubSubStruct `structmap:"myAddress"`
	Name      *string      `structmap:"name"`
	Username  string       `structmap:"user,required"`
	UserNames []string
	MyBool    bool
	MyUint    uint32
	MyFloat   float32
	MyMap     map[string]interface{}
	Headers   map[string]string `structmap:"headers"`
}

func TestDecode(t *testing.T) {
	s := &MyStruct{}
	m := map[string]interface{}{
		"name":      "Marisa",
		"user":      "{{name}}",
		"UserNames": []string{"A", "B", "C"},
		// "Age":       18,
		"Address": "Street A",
		"Number":  "1832",
		"myAddress": map[string]interface{}{
			"Address": "Street B",
			"Number":  1345,
		},
		"MyBool":  1,
		"MyUint":  true,
		"MyFloat": false,
		"MyMap": map[string]interface{}{
			"key": "value",
		},
		"headers": map[string]interface{}{
			"a": "b",
			"b": "c",
		},
	}

	defaultTag := "structmap"

	defaultValue := behavior.New(func(field *structmap.FieldPart) error {
		if field.Value != nil {
			return nil
		}

		value, _ := structmap.ParseTag(field.Tag.Get("default"))
		if value != "" {
			field.Value = value
		}

		return nil
	})

	sm := structmap.New(
		structmap.WithBehaviors(
			name.FromTag(defaultTag),
			defaultValue,
			flag.Required(defaultTag),
			cast.ToType(),
		),
	)

	if err := sm.Decode(m, s); err != nil {
		t.Error(err)
		t.FailNow()
	}

	// &{SubStruct:{SubSubStruct:0xc000010240 Age:18} MyAddress:{Address:Street B} Name:0xc0000102a0 Username:{{name}} UserNames:[A B C]}
	name := "Marisa"
	age := "15"
	n1 := 1832
	n2 := 1345

	expected := &MyStruct{
		SubStruct: SubStruct{
			SubSubStruct: &SubSubStruct{
				Address: "Street A",
				Number:  &n1,
			},
			Age: &age,
		},
		MyAddress: SubSubStruct{
			Address: "Street B",
			Number:  &n2,
		},
		Name:      &name,
		Username:  "{{name}}",
		UserNames: []string{"A", "B", "C"},
		MyBool:    true,
		MyUint:    1,
		MyFloat:   0,
		MyMap: map[string]interface{}{
			"key": "value",
		},
		Headers: map[string]string{
			"a": "b",
			"b": "c",
		},
	}

	if !reflect.DeepEqual(s, expected) {
		t.Errorf("Expected = %+v; got = %+v", expected, s)
	}
}

type DefaultTypes struct {
	Tstring    string `structmap:"tstring"`
	Tint       int    `structmap:"tint"`
	Tint8      int8   `structmap:"tint8"`
	Tint16     int16  `structmap:"tint16"`
	Tint32     int32  `structmap:"tint32"`
	Tint64     int64  `structmap:"tint64"`
	Tuint      uint
	Tbool      bool `structmap:"tbool"`
	Tfloat     float64
	unexported bool
	Tdata      interface{} `structmap:"tdata"`
}

type DefaultTypesPointer struct {
	Tstring    *string `structmap:"tstring"`
	Tint       *int    `structmap:"tint"`
	Tuint      *uint
	Tbool      *bool `structmap:"tbool"`
	Tfloat     *float64
	unexported *bool
	Tdata      *interface{} `structmap:"tdata"`
}

func TestDefaultTypes(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"tstring":    "foo",
		"tint":       20,
		"tint8":      20,
		"tint16":     20,
		"tint32":     20,
		"tint64":     20,
		"Tuint":      20,
		"tbool":      true,
		"Tfloat":     20.20,
		"unexported": true,
		"tdata":      20,
	}

	var result DefaultTypes

	sm := structmap.New(
		structmap.WithBehaviors(
			name.FromTag("structmap"),
		),
	)

	err := sm.Decode(input, &result)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if result.Tstring != "foo" {
		t.Errorf("tstring value should be 'foo': %#v", result.Tstring)
	}

	if result.Tint != 20 {
		t.Errorf("tint value should be 20: %#v", result.Tint)
	}
	if result.Tint8 != 20 {
		t.Errorf("tint8 value should be 20: %#v", result.Tint)
	}
	if result.Tint16 != 20 {
		t.Errorf("tint16 value should be 20: %#v", result.Tint)
	}
	if result.Tint32 != 20 {
		t.Errorf("tint32 value should be 20: %#v", result.Tint)
	}
	if result.Tint64 != 20 {
		t.Errorf("tint64 value should be 20: %#v", result.Tint)
	}

	if result.Tuint != 20 {
		t.Errorf("tuint value should be 20: %#v", result.Tuint)
	}

	if result.Tbool != true {
		t.Errorf("tbool value should be true: %#v", result.Tbool)
	}

	if result.Tfloat != 20.20 {
		t.Errorf("tfloat value should be 20.20: %#v", result.Tfloat)
	}

	if result.unexported != false {
		t.Error("unexported should not be set, it is unexported")
	}

	if result.Tdata != 20 {
		t.Error("tdata should be valid")
	}
}

func TestFromDefaultTypesToPointer(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"tstring":    "foo",
		"tint":       20,
		"Tuint":      20,
		"tbool":      true,
		"Tfloat":     20.20,
		"unexported": true,
		"tdata":      20,
	}

	var result DefaultTypesPointer

	sm := structmap.New(
		structmap.WithBehaviors(
			name.FromTag("structmap"),
		),
	)

	err := sm.Decode(input, &result)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if *result.Tstring != "foo" {
		t.Errorf("tstring value should be 'foo': %#v", result.Tstring)
	}

	if *result.Tint != 20 {
		t.Errorf("tint value should be 20: %#v", result.Tint)
	}

	if *result.Tuint != 20 {
		t.Errorf("tuint value should be 20: %#v", result.Tuint)
	}

	if *result.Tbool != true {
		t.Errorf("tbool value should be true: %#v", result.Tbool)
	}

	if *result.Tfloat != 20.20 {
		t.Errorf("tfloat value should be 20.20: %#v", result.Tfloat)
	}

	if result.unexported != nil {
		t.Error("unexported should not be set, it is unexported")
	}

	if *result.Tdata != 20 {
		t.Error("tdata should be valid")
	}
}

func TestFromPointerToDefaultTypes(t *testing.T) {
	t.Parallel()

	tstring := "foo"
	tint := 20
	tint8 := int8(20)
	tint16 := int16(20)
	tint32 := int32(20)
	tint64 := int64(20)
	tuint := uint(20)
	tbool := true
	tfloat := 20.20
	unexported := true
	tdata := 20

	input := map[string]interface{}{
		"tstring":    &tstring,
		"tint":       &tint,
		"tint8":      &tint8,
		"tint16":     &tint16,
		"tint32":     &tint32,
		"tint64":     &tint64,
		"Tuint":      &tuint,
		"tbool":      &tbool,
		"Tfloat":     &tfloat,
		"unexported": &unexported,
		"tdata":      &tdata,
	}

	var result DefaultTypes

	sm := structmap.New(
		structmap.WithBehaviors(
			name.FromTag("structmap"),
		),
	)

	err := sm.Decode(input, &result)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if result.Tstring != "foo" {
		t.Errorf("tstring value should be 'foo': %#v", result.Tstring)
	}

	if result.Tint != 20 {
		t.Errorf("tint value should be 20: %#v", result.Tint)
	}
	if result.Tint8 != 20 {
		t.Errorf("tint8 value should be 20: %#v", result.Tint)
	}
	if result.Tint16 != 20 {
		t.Errorf("tint16 value should be 20: %#v", result.Tint)
	}
	if result.Tint32 != 20 {
		t.Errorf("tint32 value should be 20: %#v", result.Tint)
	}
	if result.Tint64 != 20 {
		t.Errorf("tint64 value should be 20: %#v", result.Tint)
	}

	if result.Tuint != 20 {
		t.Errorf("tuint value should be 20: %#v", result.Tuint)
	}

	if result.Tbool != true {
		t.Errorf("tbool value should be true: %#v", result.Tbool)
	}

	if result.Tfloat != 20.20 {
		t.Errorf("tfloat value should be 20.20: %#v", result.Tfloat)
	}

	if result.unexported != false {
		t.Error("unexported should not be set, it is unexported")
	}

	if result.Tdata != 20 {
		t.Error("tdata should be valid")
	}
}

func TestFromPointerToPointer(t *testing.T) {
	t.Parallel()

	tstring := "foo"
	tint := 20
	tuint := uint(20)
	tbool := true
	tfloat := 20.20
	unexported := true
	tdata := 20

	input := map[string]interface{}{
		"tstring":    &tstring,
		"tint":       &tint,
		"Tuint":      &tuint,
		"tbool":      &tbool,
		"Tfloat":     &tfloat,
		"unexported": &unexported,
		"tdata":      &tdata,
	}

	var result DefaultTypesPointer

	sm := structmap.New(
		structmap.WithBehaviors(
			name.FromTag("structmap"),
		),
	)

	err := sm.Decode(input, &result)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if *result.Tstring != "foo" {
		t.Errorf("tstring value should be 'foo': %#v", result.Tstring)
	}

	if *result.Tint != 20 {
		t.Errorf("tint value should be 20: %#v", result.Tint)
	}

	if *result.Tuint != 20 {
		t.Errorf("tuint value should be 20: %#v", result.Tuint)
	}

	if *result.Tbool != true {
		t.Errorf("tbool value should be true: %#v", result.Tbool)
	}

	if *result.Tfloat != 20.20 {
		t.Errorf("tfloat value should be 20.20: %#v", result.Tfloat)
	}

	if result.unexported != nil {
		t.Error("unexported should not be set, it is unexported")
	}

	if *result.Tdata != 20 {
		t.Error("tdata should be valid")
	}
}

// TODO: Remove this test (tmp)
type testStr struct {
	Headers map[string]map[string]string
}

func TestMapCast(t *testing.T) {
	s := new(testStr)
	m := map[string]interface{}{
		"Headers": map[interface{}]interface{}{
			// "a": "1",
			// "b": "2",
			// "c": "3",
			"d": map[string]int{
				"a": 1,
			},
		},
	}

	sm := structmap.New(
		structmap.WithBehaviors(
			name.FromTag("structmap"),
			cast.ToType(),
		),
	)

	err := sm.Decode(m, s)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", s)
}

// TODO: Remove this test (tmp)
func TestName(t *testing.T) {
	s := &struct {
		FirstName string `bson:"first_name"`
		LastName  string `json:"last_name"`
		SnakeCase string
	}{}

	m := map[string]interface{}{
		"first_name": "MyFirstName",
		"last_name":  "MyLastName",
		"snake_case": "MySnakeCase",
	}

	sm := structmap.New(
		structmap.WithBehaviors(
			name.Discovery(name.FromTag("json"), name.FromTag("bson"), name.FromSnake),
		),
	)

	err := sm.Decode(m, s)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", s)
}

func TestNameNoop(t *testing.T) {
	s := &struct {
		ValueA string
		ValueB string
	}{}

	m := map[string]interface{}{
		"ValueA": "valA",
		"ValueB": "valB",
	}

	sm := structmap.New(structmap.WithBehaviors(name.Noop))

	err := sm.Decode(m, s)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", s)
}

type Person struct {
	Name string
}

type Ladies struct {
	Person `structmap:",noembedded"`
}

func TestLadies(t *testing.T) {
	s := new(Ladies)
	m := map[string]interface{}{
		"Name": "Luana",
		"Person": map[string]interface{}{
			"Name": "Jessica",
		},
	}

	sm := structmap.New(
		structmap.WithBehaviors(name.Noop, flag.NoEmbedded("structmap")),
	)

	if err := sm.Decode(m, s); err != nil {
		t.Error(err)
	}

	t.Logf("%+v", s)
}

func TestSliceToSlice(t *testing.T) {
	s := new(struct {
		Numbers []int
	})
	m := map[string]interface{}{
		"Numbers": []int{1, 2, 3},
	}

	sm := structmap.New(structmap.WithBehaviors(name.Noop))

	if err := sm.Decode(m, s); err != nil {
		t.Error(err)
	}

	t.Logf("%+v", s)
}

func TestArrayToArray(t *testing.T) {
	s := new(struct {
		Numbers [3]int
	})
	m := map[string]interface{}{
		"Numbers": [3]int{1, 2, 3},
	}

	sm := structmap.New(structmap.WithBehaviors(name.Noop))

	if err := sm.Decode(m, s); err != nil {
		t.Error(err)
	}

	t.Logf("%+v", s)
}

func TestSliceToArrayConverter(t *testing.T) {
	s := new(struct {
		Times [3]int
	})
	m := map[string]interface{}{
		"Times": []int{1588791963946, 1588791963946, 1588791963946},
	}

	sm := structmap.New(structmap.WithBehaviors(name.Noop, cast.ToType()))

	if err := sm.Decode(m, s); err != nil {
		t.Error(err)
	}

	t.Logf("%+v", s)
}

func TestSliceToArrayConverterType(t *testing.T) {
	s := new(struct {
		Times [3]time.Time
	})
	m := map[string]interface{}{
		"Times": []int{1588791963946, 1588791963946, 1588791963946},
	}

	castTime := cast.Type(time.Time{}, func(source reflect.Type, value reflect.Value) (result interface{}, err error) {
		switch cast.ToKind(value.Type()) {
		case reflect.Int:
			result = time.Unix(0, value.Int()*int64(time.Millisecond))
		case reflect.Float32:
			result = time.Unix(0, int64(value.Float())*int64(time.Millisecond))
		case reflect.String:
			result, err = time.Parse(time.RFC3339, value.String())
		default:
			err = cast.ErrNoConvertible
		}

		return
	})

	sm := structmap.New(
		structmap.WithBehaviors(
			name.Noop,
			cast.ToType(
				cast.WithTypes(
					castTime,
				),
			),
		),
	)

	if err := sm.Decode(m, s); err != nil {
		t.Error(err)
	}

	t.Logf("%+v", s)
}

func TestSimple(t *testing.T) {
	s := new(struct {
		A **string
		B struct {
			C string
		}
		D *int
	})

	var (
		a1 **string
		a2 *string
		a3 string
	)

	a3 = "A"
	a2 = &a3
	a1 = &a2

	m := map[string]interface{}{
		"A": "B",
		"B": map[string]interface{}{
			"C": a1,
		},
		"D": nil,
	}

	sm := structmap.New(structmap.WithBehaviors(name.Noop))

	if err := sm.Decode(m, s); err != nil {
		t.Error(err)
	}

	t.Logf("%+v", s)
}

func TestSetValue(t *testing.T) {
	var ptr ***string
	var val = "A"

	internal.SetValue(reflect.ValueOf(&ptr).Elem(), reflect.ValueOf(val))

	if ptr != nil {
		fmt.Println(***ptr)
	}
}

func TestCast(t *testing.T) {
	s := new(struct {
		A string
	})
	m := map[string]interface{}{
		"A": "B",
	}

	sm := structmap.New(structmap.WithBehaviors(name.Noop, cast.ToType()))

	err := sm.Decode(m, s)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", s)
}

func TestCastComplex(t *testing.T) {
	s := new(struct {
		A []*string
	})

	var (
		i0 **int
		i1 *int
		i2 int
	)

	i2 = 0
	i1 = &i2
	i0 = &i1

	m := map[string]interface{}{
		"A": []**int{i0, nil},
	}

	sm := structmap.New(structmap.WithBehaviors(name.Noop, cast.ToType()))

	err := sm.Decode(m, s)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", s)
}

func TestCastInterface(t *testing.T) {
	s := new(struct {
		A []string
	})

	m := map[string]interface{}{
		"A": []interface{}{0},
	}

	sm := structmap.New(structmap.WithBehaviors(name.Noop, cast.ToType()))

	err := sm.Decode(m, s)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", s)
}

func TestCastMap(t *testing.T) {
	s := new(struct {
		A map[string]*int
	})

	m := map[string]interface{}{
		"A": map[int]string{
			1: "2",
			2: "1",
		},
	}

	sm := structmap.New(structmap.WithBehaviors(name.Noop, cast.ToType()))

	err := sm.Decode(m, s)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(*s.A["1"])

	t.Logf("%+v", s)
}

func TestNil(t *testing.T) {
	s := new(struct {
		A *string
	})

	m := map[string]interface{}{
		"A": nil,
	}

	sm := structmap.New(structmap.WithBehaviors(name.Noop))

	err := sm.Decode(m, s)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", s)
}

type Data struct {
	Name string
	Age  int
}

func TestOverride(t *testing.T) {
	s := &Data{
		Name: "a",
		Age:  20,
	}

	m := map[string]interface{}{
		"Name": "b",
	}

	sm := structmap.New(structmap.WithBehaviors(name.Noop))

	err := sm.Decode(m, s)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", s)
}

func TestIntToBigInt(t *testing.T) {
	s := &struct {
		Number           big.Int
		NumberFromString big.Int
	}{}

	m := map[string]interface{}{
		"Number":           1000000,
		"NumberFromString": "9000000",
	}

	sm := structmap.New(
		structmap.WithBehaviors(
			name.Noop,
			cast.ToType(
				cast.WithTypes(
					cast.Type(big.Int{}, func(source reflect.Type, value reflect.Value) (result interface{}, err error) {
						switch cast.ToKind(value.Type()) {
						case reflect.String:
							bigInt := big.NewInt(0)

							if len(value.String()) > 0 {
								err = bigInt.UnmarshalText([]byte(value.String()))
							}

							result = *bigInt
						case reflect.Int:
							result = *big.NewInt(value.Int())
						default:
							err = cast.ErrNoConvertible
						}

						return
					}),
				),
			),
		),
	)

	err := sm.Decode(m, s)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%+v", s)
}
