# vim: set expandtab shiftwidth=4 :
# pylint: disable=missing-docstring

import json
import requests

from . import base
from . import settings

class KeysSymmTest(base.BaseTest):
    user = settings.EXISTING_USERS[1]
    wrong_user = settings.EXISTING_USERS[2]

    def make_put_body(self):
        return {
            'loginKey': self.user['loginKey'],
            'privateDataKey': self.user['privateDataKey'],
        }

    def get_symm_keys(self, auth=None):
        if auth is None:
            auth = self.auth_good()
        return requests.get(
            self.url_prefix(self.user) + '/keys/symm',
            headers={'content-type': 'application/json'},
            **auth)

    def put_symm_keys(self, body, auth=None):
        if auth is None:
            auth = self.auth_good()
        return requests.put(
            self.url_prefix(self.user) + '/keys/symm',
            headers={'content-type': 'application/json'},
            data=json.dumps(body),
            **auth)


    def test_get_bad_auth(self):
        resp = self.get_symm_keys(self.auth_wrong_user())
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

        resp = self.get_symm_keys(self.auth_nonexisting_user())
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

        resp = self.get_symm_keys(self.auth_bad_login_key())
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

    def test_get_success(self):
        resp = self.get_symm_keys()
        self.assertEqual(resp.status_code, requests.codes.ok)

        resp_body = json.loads(resp.text)
        self.assertFalse(resp_body.has_key('loginKey'))
        self.assertTrue(resp_body.has_key('privateDataKey'))
        self.assertEqual(resp_body['privateDataKey'], self.user['privateDataKey'])

    def test_put_bad_login_key_format(self):
        # missing loginKey
        body = self.make_put_body()
        del body['loginKey']
        resp = self.put_symm_keys(body)
        self.assertEqual(resp.status_code, requests.codes.bad_request)

        # loginKey too short
        body = self.make_put_body()
        body['loginKey'] = body['loginKey'][:-1]
        resp = self.put_symm_keys(body)
        self.assertEqual(resp.status_code, requests.codes.bad_request)

        # loginKey too long
        body = self.make_put_body()
        body['loginKey'] += '0'
        resp = self.put_symm_keys(body)
        self.assertEqual(resp.status_code, requests.codes.bad_request)

        # bad char
        body = self.make_put_body()
        body['loginKey'] = body['loginKey'][:-1] + 'x'
        resp = self.put_symm_keys(body)
        self.assertEqual(resp.status_code, requests.codes.bad_request)

        # wrong case
        body = self.make_put_body()
        body['loginKey'] = body['loginKey'][:-1] + 'A'
        resp = self.put_symm_keys(body)
        self.assertEqual(resp.status_code, requests.codes.bad_request)


    def test_put_bad_private_data_key_format(self):
        # missing privateDataKey
        body = self.make_put_body()
        del body['privateDataKey']
        resp = self.put_symm_keys(body)
        self.assertEqual(resp.status_code, requests.codes.bad_request)

        # privateDataKey too short
        body = self.make_put_body()
        body['privateDataKey'] = "A" * 43
        resp = self.put_symm_keys(body)
        self.assertEqual(resp.status_code, requests.codes.bad_request)

        # privateDataKey too long
        body = self.make_put_body()
        body['privateDataKey'] = "A" * 201
        resp = self.put_symm_keys(body)
        self.assertEqual(resp.status_code, requests.codes.bad_request)

        # bad char
        body = self.make_put_body()
        body['privateDataKey'] = body['privateDataKey'][:-1] + '%'
        resp = self.put_symm_keys(body)
        self.assertEqual(resp.status_code, requests.codes.bad_request)

    def test_put_success(self):
        original_auth = self.auth_good()
        original_login_key = self.user['loginKey']
        original_private_data_key = self.user['privateDataKey']

        # switch to new loginKey
        self.user['loginKey'] = "fedcba9876543210" * 8
        self.user['privateDataKey'] = "asdf" * 20
        body = self.make_put_body()
        resp = self.put_symm_keys(body, original_auth)
        self.assertEqual(resp.status_code, requests.codes.ok)

        # check changed privateDataKey
        self.test_get_success()

        # switch back to old loginKey
        new_auth = self.auth_good()
        self.user['loginKey'] = original_login_key
        self.user['privateDataKey'] = original_private_data_key
        body = self.make_put_body()
        resp = self.put_symm_keys(body, new_auth)
        self.assertEqual(resp.status_code, requests.codes.ok)

        # check original privateDataKey
        self.test_get_success()

