/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package validation

type ValidationError struct {
	// Stored in reverse order so that append can be used for efficient prepending
	pathCompontents []string
	err             error
}

func NewValidationError(field string, err error) *ValidationError {
	switch switchedErr := err.(type) {

	case *ValidationError:
		// err is already a ValidationError, so let's just copy it and add the field
		errCopy := *switchedErr
		errCopy.AddPathComponent(field)
		return &errCopy

	default:
		// Set the initial capacity to N=5, so that no re-slicing occurs for paths
		// no longer than N components. This should be sufficient for most use cases.
		pathCompontents := make([]string, 1, 5)
		pathCompontents[0] = field
		return &ValidationError{pathCompontents, err}
	}
}

func (ve *ValidationError) AddPathComponent(component string) {
	ve.pathCompontents = append(ve.pathCompontents, component)
}

func (ve *ValidationError) Error() string {
	path := ve.pathCompontents[0]
	for _, component := range ve.pathCompontents[1:] {
		path = component + "." + path
	}
	return "Error while processing field '" + path + "': " + ve.err.Error()
}
