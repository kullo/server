/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package validation

import (
	"errors"
	"strings"
	"testing"
)

type testdata struct {
	fieldname string
	err       error
}

func newTestdata() testdata {
	return testdata{"some_field", errors.New("Some error")}
}

func TestNewValidationError(t *testing.T) {
	data := newTestdata()
	result := NewValidationError(data.fieldname, data.err)

	if result == nil {
		t.Error("nil result")
	}

	errmsg := result.Error()
	if !strings.Contains(errmsg, data.fieldname) {
		t.Error("result doesn't contain fieldname")
	}
	if !strings.Contains(errmsg, data.err.Error()) {
		t.Error("result doesn't contain error message")
	}
}

func TestNewValidationErrorFromValidationError(t *testing.T) {
	data := newTestdata()
	inner := NewValidationError(data.fieldname, data.err)
	if inner == nil {
		t.Error("nil result (inner)")
	}
	outerFieldname := "outer"
	outer := NewValidationError(outerFieldname, inner)
	if outer == nil {
		t.Error("nil result (outer)")
	}

	errmsg := outer.Error()
	if !strings.Contains(errmsg, outerFieldname+"."+data.fieldname) {
		t.Error("result doesn't contain fieldname")
	}
	if !strings.Contains(errmsg, data.err.Error()) {
		t.Error("result doesn't contain error message")
	}
}
