package validator

import "go.mongodb.org/mongo-driver/bson/primitive"

type Validator struct {
	Errors map[string]any `json:"errors"`
}

func NewValidator() *Validator {
	return &Validator{
		Errors: make(map[string]any),
	}
}

func (v *Validator) Validate(isValid bool, key, msg string) {
	if !isValid {
		v.AddError(key, msg)
	}
}

func (v *Validator) AddError(key, msg string) {
	if _, ok := v.Errors[key]; !ok {
		v.Errors[key] = msg
	}
}

func (v *Validator) IsValid() bool {
	return len(v.Errors) == 0
}

func CheckLength(val string, min, max int) bool {
	return len(val) >= min && len(val) <= max
}

func IsEmpty(val string) bool {
	return len(val) == 0
}

func IsNotZeroInt(val int) bool {
	return val != 0
}

func IsValidID(id string) bool {
	_, err := primitive.ObjectIDFromHex(id)
	return err == nil
}
