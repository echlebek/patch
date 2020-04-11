package patch

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// Struct patches (modifies) its resource input with each message in
// patches. The resource provided must be a pointer to a struct, or an
// error will be returned.
//
// Reflection is used to traverse the struct's fields and match them with the
// keys in patches. On finding a matching field, the raw patch message is
// unmarshalled into the matching struct field.
//
// If a patch message is null, then its corresponding struct field will be set
// to its zero value.
func Struct(resource interface{}, patches map[string]*json.RawMessage) error {
	value := reflect.ValueOf(resource)
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return fmt.Errorf("can't operate on non-struct: %s", value.Kind().String())
	}
	if !value.CanAddr() {
		return errors.New("unaddressable struct value")
	}
	valueT := value.Type()
	for i := 0; i < valueT.NumField(); i++ {
		field := value.Field(i)
		if !field.CanAddr() || !field.CanInterface() {
			continue
		}
		if patch, ok := patches[jsonFieldName(valueT.Field(i))]; ok {
			if patch == nil {
				// set to zero value
				field.Set(reflect.Zero(field.Type()))
				continue
			}
			if err := json.Unmarshal([]byte(*patch), field.Addr().Interface()); err != nil {
				return err
			}
		}
	}
	return nil
}

func jsonFieldName(field reflect.StructField) string {
	name := strings.Split(field.Tag.Get("json"), ",")[0]
	if name == "" {
		name = field.Name
	}
	return name
}
