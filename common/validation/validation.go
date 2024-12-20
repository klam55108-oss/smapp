package validation

type Validatable interface {
	Validate() error
}

type Validator interface {
	Validate(Validatable) error
}

type defaultValidator struct{}

func (defaultValidator) Validate(v Validatable) error {
	return v.Validate()
}

var DefaultValidator = defaultValidator{}
