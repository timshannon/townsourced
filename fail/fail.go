// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

// failures are errors which can be returned to the client and are the result
// of user input or action in some way.  Different from an error in that they
// are used to prevent internal data from being exposed to clients, and they
// contain additional information as to the source of the error, usually their
// input sent back to them

package fail

// Fail is an error whos contents can be exposed to the client and is usually the result
// of incorrect client input
type Fail struct {
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	HTTPStatus int         `json:"-"` //gets set in the error response
}

func (f *Fail) Error() string {
	return f.Message
}

// New creates a new failure, data is optional
func New(message string, data ...interface{}) error {
	//if data is a single item don't return it as an array with one item, return it as a single item
	var fdata interface{}
	if len(data) == 1 {
		fdata = data[0]
	} else {
		fdata = data
	}

	return &Fail{
		Message:    message,
		Data:       fdata,
		HTTPStatus: 0,
	}
}

// NewFromErr returns a new failure based on the passed in error, data is optional
// if passed in error is nil, then nil is returned
func NewFromErr(err error, data ...interface{}) error {
	if err == nil {
		return nil
	}
	return New(err.Error(), data...)
}

// IsEqual tests whether an error is equal to another error / failure
func IsEqual(err, other error) bool {
	if err == nil {
		if other == nil {
			return true
		}
		return false
	}
	return err.Error() == other.Error()
}

// IsFail tests whether the passed in error is a failure
func IsFail(err error) bool {
	if err == nil {
		return false
	}
	switch err.(type) {
	case *Fail:
		return true
	default:
		return false
	}
}
