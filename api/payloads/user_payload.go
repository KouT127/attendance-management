package payloads

import (
	validation "github.com/go-ozzo/ozzo-validation/v3"
	"github.com/go-ozzo/ozzo-validation/v3/is"
)

type UserPayload struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	ImageURL string `json:"image_url"`
}

func (u *UserPayload) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.Name, validation.Required, validation.Length(1, 50)),
		validation.Field(&u.Email, validation.Required, validation.Length(1, 50), is.Email),
	)
}
