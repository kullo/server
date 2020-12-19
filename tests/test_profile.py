# vim: set expandtab shiftwidth=4 :
# pylint: disable=missing-docstring

import json
import requests

from . import base
from . import settings


class ProfileTest(base.BaseTest):
    user = settings.EXISTING_USERS[1]
    wrong_user = settings.EXISTING_USERS[2]

    def get_list(self, auth=None):
        if auth is None:
            auth = self.auth_good()
        return requests.get(
            self.url_prefix(self.user) + '/profile',
            **auth)

    def get_entry(self, key, auth=None):
        if auth is None:
            auth = self.auth_good()
        return requests.get(
            self.url_prefix(self.user) + '/profile/' + key,
            **auth)

    def modify_entry(self, key, value, last_modified, auth=None):
        if auth is None:
            auth = self.auth_good()
        return requests.put(
            self.url_prefix(self.user) + '/profile/' + key,
            params={'lastModified': last_modified},
            headers={'content-type': 'application/json'},
            data=json.dumps({'value': value}),
            **auth)

    def subtest_get_list(self, expected_entries):
        resp = self.get_list()
        self.assertEqual(resp.status_code, requests.codes.ok)
        json_result = json.loads(resp.text)
        self.assertIn('data', json_result)
        data = json_result['data']
        self.assertEqual(len(data), len(expected_entries))
        for entry, expected in zip(data, expected_entries):
            self.assertEqualOrAssign(entry, 'key', expected)
            self.assertEqualOrAssign(entry, 'value', expected)
            self.assertEqualOrAssign(entry, 'lastModified', expected)

    def subtest_get_entry(self, expected_entry):
        resp = self.get_entry(expected_entry['key'])
        self.assertEqual(resp.status_code, requests.codes.ok)
        json_result = json.loads(resp.text)
        self.assertEqualOrAssign(json_result, 'key', expected_entry)
        self.assertEqualOrAssign(json_result, 'value', expected_entry)
        self.assertEqualOrAssign(json_result, 'lastModified', expected_entry)

    def subtest_create_entry(self, entry):
        resp = self.modify_entry(entry['key'], entry['value'], 0)
        self.assertEqual(resp.status_code, requests.codes.ok)
        json_result = json.loads(resp.text)
        self.assertEqual(json_result['key'], entry['key'])
        self.assertTimestampIsNow(json_result['lastModified'])
        entry['lastModified'] = json_result['lastModified']

    def subtest_modify_entry(self, entries):
        entry = entries.pop(0)
        resp = self.modify_entry(entry['key'], entry['value'], entry['lastModified'])
        self.assertEqual(resp.status_code, requests.codes.ok)
        json_result = json.loads(resp.text)
        self.assertEqual(json_result['key'], entry['key'])
        self.assertTimestampIsNow(json_result['lastModified'])
        entry['lastModified'] = json_result['lastModified']
        entries.append(entry)

    def test_wrong_user_get_list(self):
        resp = self.get_list(auth=self.auth_good(self.wrong_user))
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

    def test_wrong_user_get_entry(self):
        resp = self.get_entry('some_key', auth=self.auth_good(self.wrong_user))
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

    def test_wrong_user_modify_entry(self):
        key = 'some_key'
        value = 'some_value'
        last_modified = 42
        resp = self.modify_entry(
            key, value, last_modified, auth=self.auth_good(self.wrong_user))
        self.assertEqual(resp.status_code, requests.codes.unauthorized)

    def test_profile(self):
        entries = []

        # get empty list
        self.subtest_get_list(entries)

        # get nonexistent entry
        resp = self.get_entry('name')
        self.assertEqual(resp.status_code, requests.codes.not_found)

        # modify nonexistent entry
        resp = self.modify_entry('name', 'John Doe', 42)
        self.assertEqual(resp.status_code, requests.codes.not_found)

        # create entry
        entries.append({'key': 'name', 'value': 'John Doe'})
        self.subtest_create_entry(entries[0])

        # get entry
        self.subtest_get_entry(entries[0])

        # create another entry
        entries.append({'key': 'organization', 'value': 'Doe Inc.'})
        self.subtest_create_entry(entries[1])

        # get nonempty list
        self.subtest_get_list(entries)

        # try to modify entry with bad last_modified
        resp = self.modify_entry(entries[0]['key'], entries[0]['value'], 42)
        self.assertEqual(resp.status_code, requests.codes.conflict)

        # verify that nothing has changed
        self.subtest_get_list(entries)

        # modify existent first entry
        self.subtest_modify_entry(entries)

        # verify the modification
        self.subtest_get_list(entries)
