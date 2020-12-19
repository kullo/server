/*
 * Copyright 2013–2020 Kullo GmbH
 *
 * This source code is licensed under the 3-clause BSD license. See LICENSE.txt
 * in the root directory of this source tree for details.
 */
package webservice

import (
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"bitbucket.org/kullo/server/challenges"
	"bitbucket.org/kullo/server/dao"
	"bitbucket.org/kullo/server/notifications"
	"bitbucket.org/kullo/server/validation"
	"github.com/emicklei/go-restful"
)

type accountsWebservice struct {
	RestfulWebService *restful.WebService
	daoUsers          *dao.Users
	daoKeysSymm       *dao.KeysSymm
	daoKeysAsymm      *dao.KeysAsymm
	localDomain       string
}

type asymmKeyPair struct {
	Pubkey  string `json:"pubkey"`
	Privkey string `json:"privkey"`
}

type registrationRecord struct {
	Address           string               `json:"address"`
	LoginKey          string               `json:"loginKey"`
	PrivateDataKey    string               `json:"privateDataKey"`
	KeypairEncryption asymmKeyPair         `json:"keypairEncryption"`
	KeypairSigning    asymmKeyPair         `json:"keypairSigning"`
	AcceptedTerms     string               `json:"acceptedTerms"`
	Challenge         challenges.Challenge `json:"challenge"`
	ChallengeAuth     string               `json:"challengeAuth"`
	ChallengeAnswer   string               `json:"challengeAnswer"`
}

type registrationData struct {
	Address        *dao.AddressesEntry
	AcceptedTerms  string
	SymmetricKeys  *dao.KeysSymmEntry
	EncryptionKeys *dao.KeysAsymmEntry
	SignatureKeys  *dao.KeysAsymmEntry
}

var ErrInvalidJson = errors.New("Invalid JSON")
var ErrInvalidChallenge = errors.New("Invalid Challenge.")

func NewAccounts(localDomain string) *accountsWebservice {
	service := &restful.WebService{}
	service.
		Path("/accounts").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	modelUsers := &dao.Users{}
	modelKeysSymm := &dao.KeysSymm{}
	modelKeysAsymm := &dao.KeysAsymm{}
	webservice := &accountsWebservice{
		RestfulWebService: service,
		daoUsers:          modelUsers,
		daoKeysSymm:       modelKeysSymm,
		daoKeysAsymm:      modelKeysAsymm,
		localDomain:       localDomain}

	// private (filtered)

	// public (unfiltered)
	service.Route(service.POST("").To(webservice.createEntry))

	return webservice
}

func (ws *accountsWebservice) createEntry(request *restful.Request, response *restful.Response) {
	regData, challengeClientAnswer, ok := ws.readEntryFromBody(request, response)
	if !ok {
		// writing error replies is done by readEntryFromBody()
		return
	}
	address := regData.Address.Address

	// get userExists and hasResetCode from DB
	userExists, err := ws.daoUsers.UserExists(address)
	if err != nil {
		writeServerError(err, response)
		return
	}

	isLocalAddress := true
	_, err = validation.ValidateLocalAddress(address, ws.localDomain)
	if err != nil {
		isLocalAddress = false
	}

	challengeOk, err := challenges.CheckChallenge(challengeClientAnswer, userExists, isLocalAddress)
	if ws.handleChallengeError(response, err) {
		return
	}

	if !challengeOk {
		// challenge is invalid => return a new challenge to the user
		ws.writeChallenge(response, address, userExists, isLocalAddress)

	} else {
		// challenge is valid => reset account or create user
		regData.Address.RegistrationCode = challengeClientAnswer.ChallengeAnswer
		language := preferredLanguage(request)
		if userExists {
			ws.handleAccountReset(response, regData, language)
		} else {
			ws.handleUserRegistration(response, regData, language)
		}
	}
}

func (ws *accountsWebservice) handleUserRegistration(response *restful.Response, regData *registrationData, language string) {
	address := regData.Address.Address
	err := ws.daoUsers.InsertEntry(regData.Address, regData.AcceptedTerms)
	switch {
	case err == dao.ErrAddressAlreadyExists:
		ws.writeUserExistsError(response)
		return
	case err != nil:
		writeServerError(err, response)
		return
	}

	err = ws.storeKeys(address, regData.SymmetricKeys, regData.EncryptionKeys, regData.SignatureKeys)
	if err != nil {
		writeServerError(err, response)
		return
	}

	writeEmptyJson(response, http.StatusOK)
	notifications.SendWelcomeMessage(address, language)
}

func (ws *accountsWebservice) handleAccountReset(response *restful.Response, regData *registrationData, language string) {
	address := regData.Address.Address
	err := ws.daoUsers.Reset(address)
	if err != nil {
		writeServerError(err, response)
		return
	}

	err = ws.storeKeys(
		regData.Address.Address,
		regData.SymmetricKeys,
		regData.EncryptionKeys,
		regData.SignatureKeys)
	if err != nil {
		writeServerError(err, response)
		return
	}

	writeEmptyJson(response, http.StatusOK)
	notifications.SendResetMessage(address, language)
}

func (ws *accountsWebservice) readEntryFromBody(
	request *restful.Request, response *restful.Response) (
	*registrationData, *challenges.ChallengeClientAnswer, bool) {

	record := &registrationRecord{}
	err := request.ReadEntity(record)
	if err != nil {
		writeRequestValidationError(response, ErrInvalidJson)
		return nil, nil, false
	}

	regData, err := ws.readRegistrationData(record)
	if err != nil {
		writeRequestValidationError(response, err)
		return nil, nil, false
	}

	challengeClientAnswer := &challenges.ChallengeClientAnswer{
		Address:         record.Address,
		Challenge:       record.Challenge,
		ChallengeAuth:   record.ChallengeAuth,
		ChallengeAnswer: record.ChallengeAnswer,
	}
	return regData, challengeClientAnswer, true
}

func (ws *accountsWebservice) readRegistrationData(record *registrationRecord) (*registrationData, error) {
	var err error
	regData := &registrationData{}

	regData.Address = &dao.AddressesEntry{}
	regData.Address.Address, err = validation.ValidateAddress(record.Address)
	if err != nil {
		return nil, validation.NewValidationError("address", err)
	}

	if len(record.AcceptedTerms) > 0 {
		regData.AcceptedTerms, err = validation.ValidateUrl(record.AcceptedTerms)
		if err != nil {
			return nil, validation.NewValidationError("acceptedTerms", err)
		}
	}

	regData.SymmetricKeys = &dao.KeysSymmEntry{}
	regData.SymmetricKeys.LoginKey, err = validation.ValidateLoginKey(record.LoginKey)
	if err != nil {
		return nil, validation.NewValidationError("loginKey", err)
	}
	regData.SymmetricKeys.PrivateDataKey, err = validation.ValidatePrivateDataKey(record.PrivateDataKey)
	if err != nil {
		return nil, validation.NewValidationError("privateDataKey", err)
	}

	regData.EncryptionKeys, err = ws.readKeysAsymmEntry(&record.KeypairEncryption, "enc")
	if err != nil {
		return nil, validation.NewValidationError("keypairEncryption", err)
	}
	regData.SignatureKeys, err = ws.readKeysAsymmEntry(&record.KeypairSigning, "sig")
	if err != nil {
		return nil, validation.NewValidationError("keypairSigning", err)
	}

	return regData, nil
}

func (ws *accountsWebservice) readKeysAsymmEntry(asymmKeyPair *asymmKeyPair, entryType string) (*dao.KeysAsymmEntry, error) {
	var err error
	keysAsymmEntry := &dao.KeysAsymmEntry{}
	keysAsymmEntry.Type = entryType
	keysAsymmEntry.ValidFrom = time.Now().Add(-15 * time.Minute)
	keysAsymmEntry.ValidUntil = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	keysAsymmEntry.Pubkey, err = validation.ValidatePublicKey(asymmKeyPair.Pubkey)
	if err != nil {
		return nil, validation.NewValidationError("pubkey", err)
	}
	keysAsymmEntry.Privkey, err = validation.ValidatePrivateKey(asymmKeyPair.Privkey)
	if err != nil {
		return nil, validation.NewValidationError("privkey", err)
	}

	return keysAsymmEntry, nil
}

func (ws *accountsWebservice) storeKeys(address string, symmKeys *dao.KeysSymmEntry,
	encKeys *dao.KeysAsymmEntry, sigKeys *dao.KeysAsymmEntry) error {

	// hash loginKey
	hash := sha512.Sum512([]byte(symmKeys.LoginKey))
	symmKeys.LoginKey = base64.StdEncoding.EncodeToString(hash[:])

	// Store symmetric keys
	err := ws.daoKeysSymm.InsertOrUpdateEntry(address, symmKeys)
	if err != nil {
		return err
	}

	// Store keypair encryption
	_, err = ws.daoKeysAsymm.InsertEntry(address, encKeys)
	if err != nil {
		return err
	}

	// Store keypair signing
	_, err = ws.daoKeysAsymm.InsertEntry(address, sigKeys)
	if err != nil {
		return err
	}

	return nil
}

func (ws *accountsWebservice) writeChallenge(
	response *restful.Response, address string,
	userExists bool, isLocalAddress bool) {

	challengeReply, err := challenges.CreateChallenge(address, userExists, isLocalAddress)
	if ws.handleChallengeError(response, err) {
		return
	}

	response.WriteHeaderAndEntity(http.StatusForbidden, challengeReply)
}

// return true iff there has been an error
func (ws *accountsWebservice) handleChallengeError(response *restful.Response, err error) bool {
	switch err {
	case nil:
		return false
	case challenges.ErrUserExists:
		ws.writeUserExistsError(response)
		return true
	default:
		writeServerError(err, response)
		return true
	}
}

func (ws *accountsWebservice) writeUserExistsError(response *restful.Response) {
	writeClientError(response, http.StatusConflict, "User already exists.")
}
