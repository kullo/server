# vim: set expandtab shiftwidth=4 :
# pylint: disable=missing-docstring

import json
import urllib2

import requests

from . import base
from . import settings

class PushGcmTest(base.BaseTest):
    user = settings.EXISTING_USERS[1]
    wrong_user = settings.EXISTING_USERS[2]

    # part before colon must have 11 chars (for Google Instance ID tokens)
    registration_token = '12345678901:123-token'

    def add_gcm_registration(self, env, token=None, user=None, auth=None):
        if token is None:
            token = self.registration_token
        if user is None:
            user = self.user
        if auth is None:
            auth = self.auth_good(user)
        body = {'registrationToken': token}
        if env != None:
            body['environment'] = env
        return requests.post(
            self.url_prefix(user) + '/push/gcm',
            headers={'content-type': 'application/json'},
            data=json.dumps(body),
            **auth)

    def remove_gcm_registration(self, token, auth=None):
        if auth is None:
            auth = self.auth_good()
        encoded_token = urllib2.quote(token, safe='')
        return requests.delete(
            self.url_prefix(self.user) + '/push/gcm/' + encoded_token,
            **auth)

    def test_add_bad_auth(self):
        resp = self.add_gcm_registration('android', auth=self.auth_good(self.wrong_user))
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

        resp = self.add_gcm_registration('android', auth=self.auth_wrong_user())
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

        resp = self.add_gcm_registration('android', auth=self.auth_nonexisting_user())
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

        resp = self.add_gcm_registration('android', auth=self.auth_bad_login_key())
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

    def test_add_failure_no_env(self):
        resp = self.add_gcm_registration(None)
        self.assertEqual(resp.status_code, requests.codes.bad_request)

    def test_add_with_bad_env(self):
        resp = self.add_gcm_registration('symbian')
        self.assertEqual(resp.status_code, requests.codes.bad_request)

    def test_add_with_good_unchanged_env(self):
        # first time (notifications_gcm record doesn't exist yet)
        resp = self.add_gcm_registration('android')
        self.assertEqual(resp.status_code, requests.codes.ok)

        # second time (notifications_gcm record already exists)
        resp = self.add_gcm_registration('android')
        self.assertEqual(resp.status_code, requests.codes.ok)

    def test_add_with_good_changed_env(self):
        # first time (notifications_gcm record doesn't exist yet)
        resp = self.add_gcm_registration('android')
        self.assertEqual(resp.status_code, requests.codes.ok)

        # second time (notifications_gcm record already exists, gets updated)
        resp = self.add_gcm_registration('ios')
        self.assertEqual(resp.status_code, requests.codes.ok)

    def test_add_with_changed_user(self):
        # first time (notifications_gcm record doesn't exist yet)
        resp = self.add_gcm_registration('android', user=self.wrong_user)
        self.assertEqual(resp.status_code, requests.codes.ok)

        # second time (notifications_gcm record already exists, gets updated)
        resp = self.add_gcm_registration('android')
        self.assertEqual(resp.status_code, requests.codes.ok)

    def test_add_with_changed_token_with_same_iid(self):
        # add first token
        resp = self.add_gcm_registration('android', token=self.registration_token)
        self.assertEqual(resp.status_code, requests.codes.ok)

        # add modified token, removing the first token
        resp = self.add_gcm_registration('android', token=self.registration_token + 'X')
        self.assertEqual(resp.status_code, requests.codes.ok)

        # removing original token fails because it has been removed
        resp = self.remove_gcm_registration(self.registration_token)
        self.assertEqual(resp.status_code, requests.codes.not_found)

    def test_remove_token_not_found(self):
        resp = self.add_gcm_registration('android')
        self.assertEqual(resp.status_code, requests.codes.ok)

        resp = self.remove_gcm_registration('token_doesnt_exist')
        self.assertEqual(resp.status_code, requests.codes.not_found)

    def test_remove_success(self):
        resp = self.add_gcm_registration('android')
        self.assertEqual(resp.status_code, requests.codes.ok)

        resp = self.remove_gcm_registration(self.registration_token)
        self.assertEqual(resp.status_code, requests.codes.ok)
