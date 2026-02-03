package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ---- Reflect struct JSON tag names ----
func jsonTagName(t reflect.Type, fieldName string) string {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Name == fieldName {
			tag := f.Tag.Get("json")
			if tag == "" {
				return strings.ToLower(f.Name)
			}
			name := strings.Split(tag, ",")[0]
			if name == "-" {
				return ""
			}
			return name
		}
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			if name := jsonTagName(f.Type, fieldName); name != "" {
				return name
			}
		}
	}
	return strings.ToLower(fieldName)
}

// getFieldType walks the StructNamespace (e.g. "Parent.Child.Grand")
// and returns the reflect.Type of the final field (exact type, not dereferenced).
// root should be reflect.TypeOf(payload) (can be pointer or struct).
func getFieldType(root reflect.Type, structNamespace string) (reflect.Type, bool) {
	t := root
	if t == nil {
		return nil, false
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, false
	}

	parts := strings.Split(structNamespace, ".")
	for i, part := range parts {
		f, ok := t.FieldByName(part)
		if !ok {
			return nil, false
		}
		if i == len(parts)-1 {
			return f.Type, true
		}

		next := f.Type
		if next.Kind() == reflect.Ptr {
			next = next.Elem()
		}
		if next.Kind() != reflect.Struct {
			return nil, false
		}
		t = next
	}
	return nil, false
}

func formatValidationErrors(structType reflect.Type, err error) map[string][]string {
	out := map[string][]string{}
	if err == nil {
		return out
	}

	if verrs, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range verrs {
			jsonName := jsonTagName(structType, fe.StructField())
			if jsonName == "" {
				jsonName = strings.ToLower(fe.StructField())
			}

			var msg string
			switch fe.Tag() {
			case "required":
				if ft, ok := getFieldType(structType, fe.StructNamespace()); ok {
					// ft is the declared type (e.g. *bool, bool, *string, etc.)
					if ft.Kind() == reflect.Ptr && ft.Elem().Kind() == reflect.Bool {
						msg = fmt.Sprintf("The %s field must be provided and set to true or false.", jsonName)
					} else if ft.Kind() == reflect.Bool {
						// non-pointer bool: required will never fail (presence not detectable), but handle gracefully
						msg = fmt.Sprintf("The %s field is required.", jsonName)
					} else {
						msg = fmt.Sprintf("The %s field is required.", jsonName)
					}
				} else {
					// Fallback if we couldn't find the field type
					msg = fmt.Sprintf("The %s field is required.", jsonName)
				}

			case "email":
				msg = fmt.Sprintf("The %s must be a valid email address.", jsonName)
			case "min":
				msg = fmt.Sprintf("The %s must be at least %s characters.", jsonName, fe.Param())
			case "gte":
				msg = fmt.Sprintf("The %s must be greater than or equal to %s.", jsonName, fe.Param())
			case "lte":
				msg = fmt.Sprintf("The %s must be less than or equal to %s.", jsonName, fe.Param())
			case "oneof":
				options := strings.ReplaceAll(fe.Param(), " ", ", ")
				msg = fmt.Sprintf("The %s must be one of the following: %s.", jsonName, options)

			case "password":
				msg = fmt.Sprintf("The %s must contain uppercase, number, and special character.", jsonName)
			default:
				msg = fmt.Sprintf("The %s is invalid (%s).", jsonName, fe.Tag())
			}

			out[jsonName] = append(out[jsonName], msg)
		}
	} else {
		out["body"] = []string{err.Error()}
	}

	return out
}

// ---- Type-safe validator middleware helper (no context needed — delivers typed value) ----
// Usage: http.HandleFunc("/users", ValidateBody(CreateUserHandler[CreateUserRequest]))
func ValidateBody[T any](handler func(w http.ResponseWriter, r *http.Request, payload T)) http.HandlerFunc {
	var validate = validator.New()
	return func(w http.ResponseWriter, r *http.Request) {
		// decode json
		var payload T
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields() // helpful: rejects unknown fields similar to strict schema
		if err := dec.Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "Invalid JSON payload",
				"errors":  map[string][]string{"body": {err.Error()}},
			})
			return
		}

		// validate
		if err := validate.Struct(payload); err != nil {
			errors := formatValidationErrors(reflect.TypeOf(payload), err)
			w.WriteHeader(422) // Unprocessable Entity
			_ = json.NewEncoder(w).Encode(map[string]any{
				"message": "Validation failed",
				"errors":  errors,
			})
			return
		}

		// success -> call handler with typed payload
		handler(w, r, payload)
	}
}
