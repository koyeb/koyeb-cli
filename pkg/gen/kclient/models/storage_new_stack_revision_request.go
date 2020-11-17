// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// StorageNewStackRevisionRequest storage new stack revision request
//
// swagger:model storageNewStackRevisionRequest
type StorageNewStackRevisionRequest struct {

	// message
	Message string `json:"message,omitempty"`

	// stack id
	StackID string `json:"stack_id,omitempty"`

	// yaml
	Yaml interface{} `json:"yaml,omitempty"`
}

// Validate validates this storage new stack revision request
func (m *StorageNewStackRevisionRequest) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *StorageNewStackRevisionRequest) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *StorageNewStackRevisionRequest) UnmarshalBinary(b []byte) error {
	var res StorageNewStackRevisionRequest
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}