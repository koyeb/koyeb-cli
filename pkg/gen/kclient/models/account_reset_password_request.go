// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// AccountResetPasswordRequest account reset password request
//
// swagger:model accountResetPasswordRequest
type AccountResetPasswordRequest struct {

	// email
	Email string `json:"email,omitempty"`
}

// Validate validates this account reset password request
func (m *AccountResetPasswordRequest) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *AccountResetPasswordRequest) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *AccountResetPasswordRequest) UnmarshalBinary(b []byte) error {
	var res AccountResetPasswordRequest
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}