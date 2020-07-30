// Code generated by go-swagger; DO NOT EDIT.

package functions

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/koyeb/koyeb-cli/pkg/kclient/models"
)

// FunctionsFetchFunctionHistoryReader is a Reader for the FunctionsFetchFunctionHistory structure.
type FunctionsFetchFunctionHistoryReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *FunctionsFetchFunctionHistoryReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewFunctionsFetchFunctionHistoryOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewFunctionsFetchFunctionHistoryBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewFunctionsFetchFunctionHistoryForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewFunctionsFetchFunctionHistoryNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		result := NewFunctionsFetchFunctionHistoryDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewFunctionsFetchFunctionHistoryOK creates a FunctionsFetchFunctionHistoryOK with default headers values
func NewFunctionsFetchFunctionHistoryOK() *FunctionsFetchFunctionHistoryOK {
	return &FunctionsFetchFunctionHistoryOK{}
}

/*FunctionsFetchFunctionHistoryOK handles this case with default header values.

A successful response.
*/
type FunctionsFetchFunctionHistoryOK struct {
	Payload *models.StorageFetchFunctionHistoryReply
}

func (o *FunctionsFetchFunctionHistoryOK) Error() string {
	return fmt.Sprintf("[GET /v1/stacks/{stack_id}/revisions/{sha}/functions/{function}/history][%d] functionsFetchFunctionHistoryOK  %+v", 200, o.Payload)
}

func (o *FunctionsFetchFunctionHistoryOK) GetPayload() *models.StorageFetchFunctionHistoryReply {
	return o.Payload
}

func (o *FunctionsFetchFunctionHistoryOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.StorageFetchFunctionHistoryReply)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewFunctionsFetchFunctionHistoryBadRequest creates a FunctionsFetchFunctionHistoryBadRequest with default headers values
func NewFunctionsFetchFunctionHistoryBadRequest() *FunctionsFetchFunctionHistoryBadRequest {
	return &FunctionsFetchFunctionHistoryBadRequest{}
}

/*FunctionsFetchFunctionHistoryBadRequest handles this case with default header values.

Validation error
*/
type FunctionsFetchFunctionHistoryBadRequest struct {
	Payload *models.CommonErrorWithFields
}

func (o *FunctionsFetchFunctionHistoryBadRequest) Error() string {
	return fmt.Sprintf("[GET /v1/stacks/{stack_id}/revisions/{sha}/functions/{function}/history][%d] functionsFetchFunctionHistoryBadRequest  %+v", 400, o.Payload)
}

func (o *FunctionsFetchFunctionHistoryBadRequest) GetPayload() *models.CommonErrorWithFields {
	return o.Payload
}

func (o *FunctionsFetchFunctionHistoryBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.CommonErrorWithFields)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewFunctionsFetchFunctionHistoryForbidden creates a FunctionsFetchFunctionHistoryForbidden with default headers values
func NewFunctionsFetchFunctionHistoryForbidden() *FunctionsFetchFunctionHistoryForbidden {
	return &FunctionsFetchFunctionHistoryForbidden{}
}

/*FunctionsFetchFunctionHistoryForbidden handles this case with default header values.

Returned when the user does not have permission to access the resource.
*/
type FunctionsFetchFunctionHistoryForbidden struct {
	Payload *models.CommonError
}

func (o *FunctionsFetchFunctionHistoryForbidden) Error() string {
	return fmt.Sprintf("[GET /v1/stacks/{stack_id}/revisions/{sha}/functions/{function}/history][%d] functionsFetchFunctionHistoryForbidden  %+v", 403, o.Payload)
}

func (o *FunctionsFetchFunctionHistoryForbidden) GetPayload() *models.CommonError {
	return o.Payload
}

func (o *FunctionsFetchFunctionHistoryForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.CommonError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewFunctionsFetchFunctionHistoryNotFound creates a FunctionsFetchFunctionHistoryNotFound with default headers values
func NewFunctionsFetchFunctionHistoryNotFound() *FunctionsFetchFunctionHistoryNotFound {
	return &FunctionsFetchFunctionHistoryNotFound{}
}

/*FunctionsFetchFunctionHistoryNotFound handles this case with default header values.

Returned when the resource does not exist.
*/
type FunctionsFetchFunctionHistoryNotFound struct {
	Payload *models.CommonError
}

func (o *FunctionsFetchFunctionHistoryNotFound) Error() string {
	return fmt.Sprintf("[GET /v1/stacks/{stack_id}/revisions/{sha}/functions/{function}/history][%d] functionsFetchFunctionHistoryNotFound  %+v", 404, o.Payload)
}

func (o *FunctionsFetchFunctionHistoryNotFound) GetPayload() *models.CommonError {
	return o.Payload
}

func (o *FunctionsFetchFunctionHistoryNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.CommonError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewFunctionsFetchFunctionHistoryDefault creates a FunctionsFetchFunctionHistoryDefault with default headers values
func NewFunctionsFetchFunctionHistoryDefault(code int) *FunctionsFetchFunctionHistoryDefault {
	return &FunctionsFetchFunctionHistoryDefault{
		_statusCode: code,
	}
}

/*FunctionsFetchFunctionHistoryDefault handles this case with default header values.

An unexpected error response
*/
type FunctionsFetchFunctionHistoryDefault struct {
	_statusCode int

	Payload *models.GatewayruntimeError
}

// Code gets the status code for the functions fetch function history default response
func (o *FunctionsFetchFunctionHistoryDefault) Code() int {
	return o._statusCode
}

func (o *FunctionsFetchFunctionHistoryDefault) Error() string {
	return fmt.Sprintf("[GET /v1/stacks/{stack_id}/revisions/{sha}/functions/{function}/history][%d] Functions_FetchFunctionHistory default  %+v", o._statusCode, o.Payload)
}

func (o *FunctionsFetchFunctionHistoryDefault) GetPayload() *models.GatewayruntimeError {
	return o.Payload
}

func (o *FunctionsFetchFunctionHistoryDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.GatewayruntimeError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}