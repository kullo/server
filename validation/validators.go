/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package validation

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var ErrBadFormat = errors.New("Bad format")
var ErrWrongDomainPart = errors.New("Wrong domain part")

var base64chars = "[0-9a-zA-Z+/=]"
var localPartMatcher = regexp.MustCompile("^[a-z0-9]+([\\.\\-_][a-z0-9]+)*$")
var domainPartMatcher = regexp.MustCompile("^([a-z0-9]+(-[a-z0-9]+)*\\.)+[a-z][a-z0-9]*(-[a-z0-9]+)*$")
var loginKeyMatcher = regexp.MustCompile("^[0-9a-f]{128}$")
var symmetricKeyMatcher = regexp.MustCompile("^" + base64chars + "{44,200}$") // min 256 bit = 32 bytes = 44 base64 chars
var pubkeyMatcher = regexp.MustCompile("^" + base64chars + "{500,}$")
var privkeyMatcher = regexp.MustCompile("^" + base64chars + "{1000,}$")

type ValidatorFunc func(string) (string, error)

func regexValidate(data string, matcher *regexp.Regexp) (string, error) {
	if !matcher.MatchString(data) {
		return "", ErrBadFormat
	}
	return data, nil
}

func addressHasValidFormat(addr string) bool {
	localAndDomainPart := strings.Split(addr, "#")
	if len(localAndDomainPart) != 2 {
		return false
	}
	if len(localAndDomainPart[0]) > 64 {
		return false
	}
	if len(localAndDomainPart[1]) > 255 {
		return false
	}
	domainLabels := strings.Split(localAndDomainPart[1], ".")
	for _, label := range domainLabels {
		if len(label) > 63 {
			return false
		}
	}
	return localPartMatcher.MatchString(localAndDomainPart[0]) &&
		domainPartMatcher.MatchString(localAndDomainPart[1])
}

func ValidateAddress(addr string) (string, error) {
	if !addressHasValidFormat(addr) {
		return "", ErrBadFormat
	}
	return addr, nil
}

func ValidateLocalAddress(addr string, localDomain string) (string, error) {
	addr, err := ValidateAddress(addr)
	if err != nil {
		return "", err
	}
	if !strings.HasSuffix(addr, "#"+localDomain) {
		return "", ErrWrongDomainPart
	}
	return addr, nil
}

func ValidateUrl(urlString string) (string, error) {
	parsed, err := url.ParseRequestURI(urlString)
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}

func ValidateLoginKey(loginKey string) (string, error) {
	return regexValidate(loginKey, loginKeyMatcher)
}

func ValidatePrivateDataKey(privateDataKey string) (string, error) {
	return regexValidate(privateDataKey, symmetricKeyMatcher)
}

func ValidatePublicKey(pubkey string) (string, error) {
	if len(pubkey) > 2000 {
		return "", ErrBadFormat
	}
	return regexValidate(pubkey, pubkeyMatcher)
}

func ValidatePrivateKey(privkey string) (string, error) {
	if len(privkey) > 8000 {
		return "", ErrBadFormat
	}
	return regexValidate(privkey, privkeyMatcher)
}
