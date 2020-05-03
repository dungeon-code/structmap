package structmap

import (
	"fmt"
	"reflect"
)

type (
	// FieldPart is a Field representation
	FieldPart struct {
		Name  string
		Value interface{}
		Type  reflect.Type
		Tag   reflect.StructTag
		Skip  bool
	}

	// MutationFunc that's change field information
	MutationFunc func(*FieldPart) error

	// Decoder is a structmap
	Decoder struct {
		mutations []MutationFunc
	}
)

// NewDecoder instance of Decoder
func NewDecoder() *Decoder {
	return &Decoder{}
}

// AddMutation a new mutation logic
func (decoder *Decoder) AddMutation(mutation MutationFunc) {
	decoder.mutations = append(decoder.mutations, mutation)
}

// Decode map to struct
func (decoder *Decoder) Decode(from map[string]interface{}, to interface{}) (err error) {
	defer func() {
		if err == nil {
			if recovered := recover(); recovered != nil {
				err = fmt.Errorf("%v", recovered)
			}
		}
	}()

	s, err := NewStruct(to)
	if err != nil {
		return err
	}

	for _, field := range s.Fields() {
		fp := &FieldPart{
			Tag:  field.Tag,
			Type: field.Type,
		}

		// run mutations
		for i, mutation := range decoder.mutations {
			if err := mutation(fp); err != nil {
				return err
			}

			// expects there first mutation get field name to get field value
			if i == 0 {
				if fp.Name == "" {
					fp.Name = field.Name
				}

				if value, ok := from[fp.Name]; ok {
					fp.Value = value
				}
			}
		}

		if fp.Skip {
			continue
		}

		if field.Type.Kind() == reflect.Struct || (field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) {
			value := field.Value

			if field.Value.Kind() == reflect.Ptr && field.IsZero() {
				pv := reflect.New(field.Type.Elem())
				field.Value.Set(pv)
			} else {
				value = field.Value.Addr()
			}

			structFrom := from

			if !field.IsEmbedded() {
				var ok bool
				if structFrom, ok = fp.Value.(map[string]interface{}); !ok {
					return fmt.Errorf("field %s cannot is a embedded struct, will expect that's value is a map[string]interface{}", fp.Name)
				}
			}

			if err := decoder.Decode(structFrom, value.Interface()); err != nil {
				return err
			}
		} else {
			value := reflect.ValueOf(fp.Value)
			fieldValue := field.Value

			// Get value element
			if value.Kind() == reflect.Ptr {
				value = value.Elem()
			}

			// Ignore if no have a value
			if value.Kind() == reflect.Invalid {
				continue
			}

			// Get field value element
			if fieldValue.Kind() == reflect.Ptr {
				if fieldValue.IsZero() {
					fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
				}

				fieldValue = fieldValue.Elem()
			}

			if value.Type().ConvertibleTo(fieldValue.Type()) {
				value = value.Convert(fieldValue.Type())
			}

			if value.Kind() != fieldValue.Kind() {
				return fmt.Errorf("field %s value of type %s is not assignable to type %s", field.Name, value.Type(), fieldValue.Type())
			}

			if field.Value.Kind() == reflect.Ptr {
				field.Value.Elem().Set(value)
			} else {
				field.Value.Set(value)
			}
		}
	}

	return nil
}
