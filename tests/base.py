# vim: set expandtab shiftwidth=4 :
# pylint: disable=missing-docstring

from datetime import datetime
import unittest
import urllib2

from . import settings

VALUE_NOT_AVAILABLE = 'not available but also not a failure'

class BaseTest(unittest.TestCase):
    user = None
    wrong_user = None

    @classmethod
    def url_prefix(cls, user):
        return settings.SERVER + '/' + urllib2.quote(user['address'])

    @classmethod
    def extract_id_lastmodified(cls, entry):
        return {'id': entry['id'], 'lastModified': entry['lastModified']}

    def assertTimestampIsNow(self, timestamp):
        delta = datetime.utcnow() - datetime.utcfromtimestamp(timestamp / 1e6)
        self.assertTrue(abs(delta) <= settings.ALLOWED_CLOCK_DIFFERENCE)

    def assertIsoTimeIsNow(self, isotime):
        delta = datetime.utcnow() - datetime.strptime(isotime, '%Y-%m-%dT%H:%M:%SZ')
        self.assertTrue(abs(delta) <= settings.ALLOWED_CLOCK_DIFFERENCE)

    def assertEqualIfAvailable(self, given, expected):
        if expected != VALUE_NOT_AVAILABLE:
            self.assertEqual(given, expected)

    def assertEqualOrAssign(self, source, key, dest):
        if dest[key] == VALUE_NOT_AVAILABLE:
            dest[key] = source[key]
        else:
            self.assertEqual(source[key], dest[key])

    def auth_good(self, user=None):
        if user is None:
            user = self.user
        return {'auth': (user['address'], user['loginKey'])}

    def auth_wrong_user(self):
        return {'auth': (self.wrong_user['address'], self.user['loginKey'])}

    def auth_nonexisting_user(self):
        return {'auth': ('doesntexist#kullo.test', self.user['loginKey'])}

    def auth_bad_login_key(self):
        return {'auth': (self.user['address'], 'baadbaad' * 16)}

