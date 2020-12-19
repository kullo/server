# vim: set expandtab shiftwidth=4 :
# pylint: disable=missing-docstring

import json
import requests

from . import base
from . import settings

class AccountTest(base.BaseTest):
    user = settings.EXISTING_USERS[1]
    wrong_user = settings.EXISTING_USERS[2]

    def get_info(self, auth=None, user=None, languages=None):
        if user is None:
            user = self.user
        if auth is None:
            auth = self.auth_good(user)
        headers = {}
        if languages is not None:
            headers['Accept-Language'] = languages
        return requests.get(
            self.url_prefix(user) + '/account/info',
            headers=headers,
            **auth)


    def test_get_bad_auth(self):
        resp = self.get_info(self.auth_wrong_user())
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

        resp = self.get_info(self.auth_nonexisting_user())
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

        resp = self.get_info(self.auth_bad_login_key())
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

    def test_get_success_free(self):
        resp = self.get_info()
        self.assertEqual(resp.status_code, requests.codes.ok)

        resp_body = json.loads(resp.text)
        self.assertTrue(resp_body['settingsLocation'].startswith('https://'))
        self.assertEqual(resp_body['planName'], 'Free')
        self.assertEqual(resp_body['storageQuota'], settings.STORAGE_QUOTA[resp_body['planName']])
        self.assertTrue('storageUsed' in resp_body)

    def test_get_success_friend(self):
        resp = self.get_info(user=self.wrong_user)
        self.assertEqual(resp.status_code, requests.codes.ok)

        resp_body = json.loads(resp.text)
        self.assertTrue(resp_body['settingsLocation'].startswith('https://'))
        self.assertEqual(resp_body['planName'], 'Friend')
        self.assertEqual(resp_body['storageQuota'], settings.STORAGE_QUOTA[resp_body['planName']])
        self.assertTrue('storageUsed' in resp_body)

    def test_language_only(self):
        resp = self.get_info(languages='de')
        self.assertEqual(resp.status_code, requests.codes.ok)

    def test_language_with_country(self):
        resp = self.get_info(languages='de-DE')
        self.assertEqual(resp.status_code, requests.codes.ok)

    def test_language_invalid(self):
        resp = self.get_info(languages='ork')
        # doesn't fail but fall back internally on 'en'
        self.assertEqual(resp.status_code, requests.codes.ok)
