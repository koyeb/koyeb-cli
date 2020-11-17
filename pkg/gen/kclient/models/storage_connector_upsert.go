// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// StorageConnectorUpsert storage connector upsert
//
// swagger:model storageConnectorUpsert
type StorageConnectorUpsert struct {

	// Only valid for update (whether or not to regen the url)
	ChangeURL bool `json:"change_url,omitempty"`

	// Name of the connector
	Name string `json:"name,omitempty"`

	// Cloudevent webhook metadata
	WebhookCloudevent *StorageMetadataWebhookCloudEvent `json:"webhook_cloudevent,omitempty"`

	// RawHttp webhook metadata
	WebhookRawhttp *StorageMetadataWebhookRawHTTP `json:"webhook_rawhttp,omitempty"`
}

// Validate validates this storage connector upsert
func (m *StorageConnectorUpsert) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateWebhookCloudevent(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateWebhookRawhttp(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *StorageConnectorUpsert) validateWebhookCloudevent(formats strfmt.Registry) error {

	if swag.IsZero(m.WebhookCloudevent) { // not required
		return nil
	}

	if m.WebhookCloudevent != nil {
		if err := m.WebhookCloudevent.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("webhook_cloudevent")
			}
			return err
		}
	}

	return nil
}

func (m *StorageConnectorUpsert) validateWebhookRawhttp(formats strfmt.Registry) error {

	if swag.IsZero(m.WebhookRawhttp) { // not required
		return nil
	}

	if m.WebhookRawhttp != nil {
		if err := m.WebhookRawhttp.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("webhook_rawhttp")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *StorageConnectorUpsert) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *StorageConnectorUpsert) UnmarshalBinary(b []byte) error {
	var res StorageConnectorUpsert
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}