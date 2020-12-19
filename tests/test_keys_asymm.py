# vim: set expandtab shiftwidth=4 :
# pylint: disable=missing-docstring

import json
import requests

from . import base
from . import settings

class KeysAsymmTest(base.BaseTest):
    user = settings.EXISTING_USERS[1]
    wrong_user = settings.EXISTING_USERS[2]

    def get_pub_keys(self, auth=None):
        if auth is None:
            auth = self.auth_good()
        return requests.get(
            self.url_prefix(self.user) + '/keys/public',
            headers={'content-type': 'application/json'},
            **auth)

    def get_pub_key(self, key_id, auth=None):
        if auth is None:
            auth = self.auth_good()
        return requests.get(
            self.url_prefix(self.user) + '/keys/public/' + unicode(key_id),
            headers={'content-type': 'application/json'},
            **auth)

    def get_priv_keys(self, auth=None):
        if auth is None:
            auth = self.auth_good()
        return requests.get(
            self.url_prefix(self.user) + '/keys/private',
            headers={'content-type': 'application/json'},
            **auth)

    def get_priv_key(self, key_id, auth=None):
        if auth is None:
            auth = self.auth_good()
        return requests.get(
            self.url_prefix(self.user) + '/keys/private/' + unicode(key_id),
            headers={'content-type': 'application/json'},
            **auth)


    def test_get_priv_keys_bad_auth(self):
        resp = self.get_priv_keys(self.auth_wrong_user())
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

        resp = self.get_priv_keys(self.auth_nonexisting_user())
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

        resp = self.get_priv_keys(self.auth_bad_login_key())
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

    def test_get_priv_key_bad_auth(self):
        resp = self.get_priv_key(1, self.auth_wrong_user())
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

        resp = self.get_priv_key(1, self.auth_nonexisting_user())
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

        resp = self.get_priv_key(1, self.auth_bad_login_key())
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

    def test_get_pub_keys_success(self):
        resp = self.get_pub_keys()
        self.assertEqual(resp.status_code, requests.codes.ok)

        resp_body = json.loads(resp.text)
        for key in resp_body:
            self.assertTrue(key.has_key('id'))
            self.assertTrue(key.has_key('type'))
            self.assertTrue(key.has_key('pubkey'))
            self.assertFalse(key.has_key('privkey'))
            self.assertTrue(key.has_key('validFrom'))
            self.assertTrue(key.has_key('validUntil'))
            self.assertTrue(key.has_key('revocation'))

    def test_get_pub_key_success(self):
        resp = self.get_pub_key(1)
        self.assertEqual(resp.status_code, requests.codes.ok)

        resp_body = json.loads(resp.text)
        self.assertTrue(resp_body.has_key('id'))
        self.assertTrue(resp_body.has_key('type'))
        self.assertTrue(resp_body.has_key('pubkey'))
        self.assertFalse(resp_body.has_key('privkey'))
        self.assertTrue(resp_body.has_key('validFrom'))
        self.assertTrue(resp_body.has_key('validUntil'))
        self.assertTrue(resp_body.has_key('revocation'))

    def test_get_priv_keys_success(self):
        resp = self.get_priv_keys()
        self.assertEqual(resp.status_code, requests.codes.ok)

        resp_body = json.loads(resp.text)
        for key in resp_body:
            self.assertTrue(key.has_key('id'))
            self.assertTrue(key.has_key('type'))
            self.assertTrue(key.has_key('pubkey'))
            self.assertTrue(key.has_key('privkey'))
            self.assertTrue(key.has_key('validFrom'))
            self.assertTrue(key.has_key('validUntil'))
            self.assertTrue(key.has_key('revocation'))

    def test_get_priv_key_success(self):
        resp = self.get_priv_key(1)
        self.assertEqual(resp.status_code, requests.codes.ok)

        resp_body = json.loads(resp.text)
        self.assertTrue(resp_body.has_key('id'))
        self.assertTrue(resp_body.has_key('type'))
        self.assertTrue(resp_body.has_key('pubkey'))
        self.assertTrue(resp_body.has_key('privkey'))
        self.assertTrue(resp_body.has_key('validFrom'))
        self.assertTrue(resp_body.has_key('validUntil'))
        self.assertTrue(resp_body.has_key('revocation'))

