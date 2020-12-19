/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package logging

import (
	"net/url"
	"testing"
)

func TestCensorRoot(t *testing.T) {
	root, err := url.Parse("http://localhost/")
	if err != nil {
		t.Fatal("url.Parse")
	}
	result := censor(root)
	if result != "/" {
		t.Error("Result is", result)
	}
}

func TestCensorStatus(t *testing.T) {
	root, err := url.Parse("http://localhost/status/")
	if err != nil {
		t.Fatal("url.Parse")
	}
	result := censor(root)
	if result != "/status/" {
		t.Error("Result is", result)
	}
}

func TestCensorAddress(t *testing.T) {
	root, err := url.Parse("http://localhost/foo%23kullo.net/messages/")
	if err != nil {
		t.Fatal("url.Parse")
	}
	result := censor(root)
	if result != "/<address>/messages/" {
		t.Error("Result is", result)
	}
}
