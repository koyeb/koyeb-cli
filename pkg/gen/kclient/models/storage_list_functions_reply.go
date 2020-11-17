// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// StorageListFunctionsReply storage list functions reply
//
// swagger:model storageListFunctionsReply
type StorageListFunctionsReply struct {

	// functions
	Functions []*StorageFunctionListItem `json:"functions"`
}

// Validate validates this storage list functions reply
func (m *StorageListFunctionsReply) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateFunctions(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *StorageListFunctionsReply) validateFunctions(formats strfmt.Registry) error {

	if swag.IsZero(m.Functions) { // not required
		return nil
	}

	for i := 0; i < len(m.Functions); i++ {
		if swag.IsZero(m.Functions[i]) { // not required
			continue
		}

		if m.Functions[i] != nil {
			if err := m.Functions[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("functions" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *StorageListFunctionsReply) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *StorageListFunctionsReply) UnmarshalBinary(b []byte) error {
	var res StorageListFunctionsReply
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}