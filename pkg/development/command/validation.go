package command

import (
	"fmt"
	"regexp"

	refdocker "github.com/containerd/containerd/reference/docker"
	jvalidator "github.com/go-playground/validator/v10"
	"k8s.io/apimachinery/pkg/api/resource"
)

type ErrorResponse struct {
	FailedField string
	Tag         string
	Value       string
}

func validateRequiredCPU(fl jvalidator.FieldLevel) bool {
	value := fl.Field().String()

	return validateQuantity(value)
}

func validateRequiredMemory(fl jvalidator.FieldLevel) bool {
	value := fl.Field().String()

	return validateQuantity(value)
}

func validateLimitedCPU(fl jvalidator.FieldLevel) bool {
	value := fl.Field().String()
	// limitedCPU is optional field
	if value == "" {
		return true
	}
	return validateQuantity(value)
}

func validateLimitedMemory(fl jvalidator.FieldLevel) bool {
	value := fl.Field().String()
	// limitedMemory is optional field
	if value == "" {
		return true
	}
	return validateQuantity(value)
}

func validateName(fl jvalidator.FieldLevel) bool {
	value := fl.Field().String()
	//if errs := validation.IsDNS1123Label(value); len(errs) > 0 {
	//	return false
	//}
	match, _ := regexp.MatchString("^[a-z0-9]{1,30}$", value)
	return match
}

func validateImage(fl jvalidator.FieldLevel) bool {
	value := fl.Field().String()
	_, err := refdocker.ParseDockerRef(value)
	if err != nil {
		return false
	}
	return true
}

func validateQuantity(value string) bool {
	_, err := resource.ParseQuantity(value)
	if err != nil {
		return false
	}
	return true
}
func init() {
	validate.RegisterValidation("requiredCpu", validateRequiredCPU)
	validate.RegisterValidation("requiredMemory", validateRequiredMemory)
	validate.RegisterValidation("limitedCpu", validateLimitedCPU)
	validate.RegisterValidation("limitedMemory", validateLimitedMemory)
	validate.RegisterValidation("name", validateName)
	validate.RegisterValidation("image", validateImage)
}

func ValidateStruct(data interface{}) []ErrorResponse {
	var errs []ErrorResponse
	err := validate.Struct(data)
	if err != nil {
		for _, e := range err.(jvalidator.ValidationErrors) {
			var elem ErrorResponse
			elem.FailedField = e.StructNamespace()
			elem.Tag = e.Tag()
			elem.Value = fmt.Sprintf("%v", e.Value())
			errs = append(errs, elem)
		}
	}
	return errs
}
