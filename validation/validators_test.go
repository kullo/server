/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package validation

import (
	"strings"
	"testing"
)

func expectValid(t *testing.T, fun ValidatorFunc, arg string, errmsg string) {
	result, err := fun(arg)
	if err != nil || result != arg {
		t.Error(errmsg)
	}
}

func expectInvalid(t *testing.T, fun ValidatorFunc, arg string, errmsg string) {
	result, err := fun(arg)
	if err == nil || result != "" {
		t.Error(errmsg)
	}
}

func validateKulloDotNetAddress(addr string) (string, error) {
	return ValidateLocalAddress(addr, "kullo.net")
}

func TestAddressHashes(t *testing.T) {
	expectInvalid(t, ValidateAddress, "testkullo.net", "no hash")
	expectValid(t, ValidateAddress, "test#kullo.net", "one hash")
	expectInvalid(t, ValidateAddress, "test#kullo#net", "two hashes")
}

func TestAddressBadChars(t *testing.T) {
	expectInvalid(t, ValidateAddress, "usr(name#kullo.net", "bad char (")
	expectInvalid(t, ValidateAddress, "usr)name#kullo.net", "bad char )")
	expectInvalid(t, ValidateAddress, "usr<name#kullo.net", "bad char <")
	expectInvalid(t, ValidateAddress, "usr>name#kullo.net", "bad char >")
	expectInvalid(t, ValidateAddress, "usr[name#kullo.net", "bad char [")
	expectInvalid(t, ValidateAddress, "usr]name#kullo.net", "bad char ]")
	expectInvalid(t, ValidateAddress, "usr{name#kullo.net", "bad char {")
	expectInvalid(t, ValidateAddress, "usr}name#kullo.net", "bad char }")
	expectInvalid(t, ValidateAddress, "usr|name#kullo.net", "bad char |")
	expectInvalid(t, ValidateAddress, "usr&name#kullo.net", "bad char &")
	expectInvalid(t, ValidateAddress, "usr!name#kullo.net", "bad char !")
	expectInvalid(t, ValidateAddress, "usr?name#kullo.net", "bad char ?")
	expectInvalid(t, ValidateAddress, "usr^name#kullo.net", "bad char ^")
	expectInvalid(t, ValidateAddress, "usr&name#kullo.net", "bad char &")
	expectInvalid(t, ValidateAddress, "usr%name#kullo.net", "bad char %")
	expectInvalid(t, ValidateAddress, "usr$name#kullo.net", "bad char $")
	expectInvalid(t, ValidateAddress, "usr*name#kullo.net", "bad char *")
	expectInvalid(t, ValidateAddress, "usr+name#kullo.net", "bad char +")
	expectInvalid(t, ValidateAddress, "usr=name#kullo.net", "bad char =")
	expectInvalid(t, ValidateAddress, "usr~name#kullo.net", "bad char ~")
	expectInvalid(t, ValidateAddress, "usr`name#kullo.net", "bad char `")
	expectInvalid(t, ValidateAddress, "usr'name#kullo.net", "bad char '")
	expectInvalid(t, ValidateAddress, "usr:name#kullo.net", "bad char :")
	expectInvalid(t, ValidateAddress, "usr;name#kullo.net", "bad char ;")
	expectInvalid(t, ValidateAddress, "usr@name#kullo.net", "bad char @")
	expectInvalid(t, ValidateAddress, "usr/name#kullo.net", "bad char /")
	expectInvalid(t, ValidateAddress, "usr\\name#kullo.net", "bad char \\")
	expectInvalid(t, ValidateAddress, "usr,name#kullo.net", "bad char ,")
	expectInvalid(t, ValidateAddress, "usr\"name#kullo.net", "bad char \"")
}

func TestAddressLocalPartSeparators(t *testing.T) {
	expectValid(t, ValidateAddress, "user.name#kullo.net", "local part with separator .")
	expectValid(t, ValidateAddress, "user-name#kullo.net", "local part with separator -")
	expectValid(t, ValidateAddress, "user_name#kullo.net", "local part with separator _")

	expectInvalid(t, ValidateAddress, ".name#kullo.net", "local part starting with .")
	expectInvalid(t, ValidateAddress, "-name#kullo.net", "local part starting with -")
	expectInvalid(t, ValidateAddress, "_name#kullo.net", "local part starting with _")

	expectInvalid(t, ValidateAddress, "name.#kullo.net", "local part ending with .")
	expectInvalid(t, ValidateAddress, "name-#kullo.net", "local part ending with -")
	expectInvalid(t, ValidateAddress, "name_#kullo.net", "local part ending with _")

	expectInvalid(t, ValidateAddress, "user..name#kullo.net", "local part with ..")
	expectInvalid(t, ValidateAddress, "user.-name#kullo.net", "local part with .-")
	expectInvalid(t, ValidateAddress, "user._name#kullo.net", "local part with ._")
	expectInvalid(t, ValidateAddress, "user-.name#kullo.net", "local part with -.")
	expectInvalid(t, ValidateAddress, "user--name#kullo.net", "local part with --")
	expectInvalid(t, ValidateAddress, "user-_name#kullo.net", "local part with -_")
	expectInvalid(t, ValidateAddress, "user_.name#kullo.net", "local part with _.")
	expectInvalid(t, ValidateAddress, "user_-name#kullo.net", "local part with _-")
	expectInvalid(t, ValidateAddress, "user__name#kullo.net", "local part with __")
}

func TestAddressDomain(t *testing.T) {
	expectInvalid(t, ValidateAddress, "test#.kullo.net", "domain with leading .")
	expectInvalid(t, ValidateAddress, "test#kullo..net", "domain with double ..")
	expectInvalid(t, ValidateAddress, "test#kullo.net.", "domain with trailing .")
	expectInvalid(t, ValidateAddress, "test#kullo", "address without .")

	expectValid(t, ValidateAddress, "test#ku-llo.net", "domain with -")
	expectValid(t, ValidateAddress, "test#ku.llo.net", "domain with two .")
	expectInvalid(t, ValidateAddress, "test#-kullo.net", "domain with leading -")
	expectInvalid(t, ValidateAddress, "test#kullo-.net", "domain with trailing -")
	expectInvalid(t, ValidateAddress, "test#ku--llo.net", "domain with double -")
	expectInvalid(t, ValidateAddress, "test#kullo.-net", "tld with leading -")
	expectInvalid(t, ValidateAddress, "test#kullo.net-", "tld with trailing -")
	expectInvalid(t, ValidateAddress, "test#kullo.123", "numeric tld")
}

func TestAddressChars(t *testing.T) {
	expectValid(t, ValidateAddress, "abcdefghijklmnopqrstuvwxyz#kullo.net", "local part: lowercase alpha")
	expectInvalid(t, ValidateAddress, "ABCDEFGHIJKLMNOPQRSTUVWXYZ#kullo.net", "local part: uppercase alpha")
	expectValid(t, ValidateAddress, "0123456789#kullo.net", "local part: numeric")

	expectValid(t, ValidateAddress, "test#abcdefghijklmnopqr.stuvwxyz", "domain part: lowercase alpha")
	expectInvalid(t, ValidateAddress, "test#ABCDEFGHIJKLMNOPQR.STUVWXYZ", "domain part: uppercase alpha")
	expectValid(t, ValidateAddress, "test#0123456789.net", "domain part: numeric")
}

func TestAddressLength(t *testing.T) {
	expectValid(t, ValidateAddress, "a#kullo.net", "local part: 1 char")
	expectValid(t, ValidateAddress, "a#b.c", "shortest possible")
	expectInvalid(t, ValidateAddress, "", "0 chars")
	expectInvalid(t, ValidateAddress, "a#", "domain part: 0 chars")
	expectInvalid(t, ValidateAddress, "#kullo.net", "local part: 0 chars")

	maxUsername := strings.Repeat("a", 64)
	maxDomainLabel := strings.Repeat("b", 63)
	maxDomain := maxDomainLabel + "." + maxDomainLabel + "." +
		maxDomainLabel + "." + maxDomainLabel
	expectValid(t, ValidateAddress, maxUsername+"#"+maxDomain, "longest possible")
	expectInvalid(t, ValidateAddress, maxUsername+"a#"+maxDomain, "local part too long")
	expectInvalid(t, ValidateAddress, maxUsername+"#b."+maxDomain, "domain part too long")
	expectInvalid(t, ValidateAddress, maxUsername+"#b.b"+maxDomainLabel, "domain label too long")
}

func TestLocalAddress(t *testing.T) {
	expectInvalid(t, validateKulloDotNetAddress, "invalid@kullo.net", "invalid address")
	expectInvalid(t, validateKulloDotNetAddress, "someone#example.net", "non-local address")

	expectValid(t, validateKulloDotNetAddress, "someone#kullo.net", "local address")
}

func TestLoginKey(t *testing.T) {
	expectInvalid(t, ValidateLoginKey, "", "empty")
	expectInvalid(t, ValidateLoginKey, strings.Repeat("0123456789Abcdef", 128/16), "uppercase")
	expectInvalid(t, ValidateLoginKey, strings.Repeat("0123456789bcdefg", 128/16), "bad char")

	expectValid(t, ValidateLoginKey, strings.Repeat("0123456789abcdef", 128/16), "valid")
}

var allBase64chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+/="

func cutOrRepeat(str string, length int) string {
	result := strings.Repeat(str, length/len(str))
	result += str[:length%len(str)]
	return result
}

func TestPrivateDataKey(t *testing.T) {
	expectInvalid(t, ValidatePrivateDataKey, "", "empty")
	expectInvalid(t, ValidatePrivateDataKey, allBase64chars[:43], "too short")
	expectInvalid(t, ValidatePrivateDataKey, allBase64chars[:43]+"-", "bad char")
	expectInvalid(t, ValidatePrivateDataKey, cutOrRepeat(allBase64chars, 1000), "too long")

	expectValid(t, ValidatePrivateDataKey, allBase64chars[:44], "valid (first part)")
	expectValid(t, ValidatePrivateDataKey, allBase64chars[len(allBase64chars)-44:], "valid (second part)")
	expectValid(t, ValidatePrivateDataKey, cutOrRepeat(allBase64chars, 200), "valid (long)")
}

func TestPublicKey(t *testing.T) {
	expectInvalid(t, ValidatePublicKey, "", "empty")
	expectInvalid(t, ValidatePublicKey, cutOrRepeat(allBase64chars, 499), "too short")
	expectInvalid(t, ValidatePublicKey, cutOrRepeat(allBase64chars, 499)+"-", "bad char")
	expectInvalid(t, ValidatePublicKey, cutOrRepeat(allBase64chars, 4000), "too long")

	expectValid(t, ValidatePublicKey, cutOrRepeat(allBase64chars, 500), "valid (shortest)")
	expectValid(t, ValidatePublicKey, cutOrRepeat(allBase64chars, 2000), "valid (long)")
}

func TestPrivateKey(t *testing.T) {
	expectInvalid(t, ValidatePublicKey, "", "empty")
	expectInvalid(t, ValidatePrivateKey, cutOrRepeat(allBase64chars, 999), "too short")
	expectInvalid(t, ValidatePrivateKey, cutOrRepeat(allBase64chars, 999)+"-", "bad char")
	expectInvalid(t, ValidatePrivateKey, cutOrRepeat(allBase64chars, 10000), "too long")

	expectValid(t, ValidatePrivateKey, cutOrRepeat(allBase64chars, 1000), "valid (shortest)")
	expectValid(t, ValidatePrivateKey, cutOrRepeat(allBase64chars, 4000), "valid (long)")
}
