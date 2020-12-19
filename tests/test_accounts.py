# vim: set expandtab shiftwidth=4 :
# pylint: disable=missing-docstring

import json
import requests

from . import base
from . import db
from . import settings

def make_body(user):
    return {
        'address': user['address'],
        'loginKey': user['loginKey'],
        'privateDataKey': user['privateDataKey'],
        'keypairEncryption': {
            'pubkey': user['encryptionPubkey'],
            'privkey': user['encryptionPrivkey'],
        },
        'keypairSigning': {
            'pubkey': user['signingPubkey'],
            'privkey': user['signingPrivkey'],
        },
        'acceptedTerms': user['acceptedTerms'],
    }

def register_account(body, languages=None):
    headers = {'content-type': 'application/json'}
    if languages is not None:
        headers['Accept-Language'] = languages
    return requests.post(
        settings.SERVER + '/accounts',
        headers=headers,
        data=json.dumps(body))

def update_body_with_challenge(req_body, resp_body):
    req_body['challenge'] = resp_body['challenge']
    req_body['challengeAuth'] = resp_body['challengeAuth']


class AccountsTest(base.BaseTest):
    def tearDown(self):
        with db.get_connection(settings.DB_CONNECTION_STRING) as conn:
            with conn.cursor() as cursor:
                for user in settings.NONEXISTING_USERS.itervalues():
                    db.delete_user(cursor, user)
                for user in settings.RESERVATION_USERS.itervalues():
                    db.delete_user(cursor, user)

    def send_initial_request(
            self, user, expected_challenge_type,
            expected_error_code=requests.codes.forbidden):

        req_body = make_body(user)
        resp = register_account(req_body)
        self.assertEqual(resp.status_code, expected_error_code)
        resp_body = json.loads(resp.text)
        if expected_challenge_type is not None:
            self.assertEqual(
                resp_body['challenge']['type'],
                expected_challenge_type)
        return req_body, resp_body


    def test_fail_on_existing_user(self):
        # send initial request with existing user
        self.send_initial_request(
            settings.EXISTING_USERS[1],
            None,
            requests.codes.conflict)

    def test_fail_on_inconsistent_user(self):
        user = settings.RESERVATION_USERS[1]

        # send initial request
        req_body, resp_body = self.send_initial_request(user, 'reservation')

        # reply with correct answer but modified address
        update_body_with_challenge(req_body, resp_body)
        req_body['challengeAnswer'] = user['reservation']
        req_body['address'] = settings.RESERVATION_USERS[2]['address']
        resp = register_account(req_body)
        self.assertEqual(resp.status_code, requests.codes.forbidden)

    def test_fail_on_modified_challenge(self):
        user = settings.RESERVATION_USERS[1]

        # send initial request
        req_body, resp_body = self.send_initial_request(user, 'reservation')

        # reply with correct answer but modified challenge
        for field, value in (
                ('type', 'bad'),
                ('user', 'bad#kullo.test'),
                ('timestamp', 1234567890),
                ('text', 'bad')):
            update_body_with_challenge(req_body, resp_body)
            req_body['challengeAnswer'] = user['reservation']
            req_body['challenge'][field] = value
            resp = register_account(req_body)
            self.assertEqual(resp.status_code, requests.codes.forbidden)

        # reply with correct answer but modified challenge auth
        update_body_with_challenge(req_body, resp_body)
        req_body['challengeAnswer'] = user['reservation']
        req_body['challengeAuth'] = 'bad'
        resp = register_account(req_body)
        self.assertEqual(resp.status_code, requests.codes.forbidden)

    def test_reservation_fail_on_wrong_answer(self):
        user = settings.RESERVATION_USERS[1]

        # send initial request for reservation user
        req_body, resp_body = self.send_initial_request(user, 'reservation')

        # reply with wrong answer
        update_body_with_challenge(req_body, resp_body)
        req_body['challengeAnswer'] = 'bad'
        resp = register_account(req_body)
        self.assertEqual(resp.status_code, requests.codes.forbidden)

    def test_reservation_success(self):
        user = settings.RESERVATION_USERS[1]

        # send initial request for reservation user
        req_body, resp_body = self.send_initial_request(user, 'reservation')

        # reply with correct answer
        update_body_with_challenge(req_body, resp_body)
        req_body['challengeAnswer'] = user['reservation']
        resp = register_account(req_body, languages='de-DE')
        self.assertEqual(resp.status_code, requests.codes.ok)

        #TODO check user inbox

    def test_fail_on_nonlocal_non_preregistered_address(self):
        user = settings.NONLOCAL_USERS[1]

        # send initial request for reservation user
        req_body, resp_body = self.send_initial_request(user, 'blocked')

    def test_reservation_success_with_nonlocal_address(self):
        user = settings.NONLOCAL_RESERVATION_USERS[1]

        # send initial request for reservation user
        req_body, resp_body = self.send_initial_request(user, 'reservation')

        # reply with correct answer
        update_body_with_challenge(req_body, resp_body)
        req_body['challengeAnswer'] = user['reservation']
        resp = register_account(req_body)
        self.assertEqual(resp.status_code, requests.codes.ok)

        #TODO check user inbox

    def test_reset_fail_on_wrong_answer(self):
        user = settings.RESET_USERS[1]

        #TODO add some messages

        # send initial request for reset user
        req_body, resp_body = self.send_initial_request(user, 'reset')

        # reply with wrong answer
        update_body_with_challenge(req_body, resp_body)
        req_body['challengeAnswer'] = 'bad'
        resp = register_account(req_body)
        self.assertEqual(resp.status_code, requests.codes.forbidden)

        #TODO check that old login still works
        #TODO check that old messages are still there

    def test_reset_success(self):
        user = settings.RESET_USERS[1]

        #TODO add some messages

        # send initial request for reset user
        req_body, resp_body = self.send_initial_request(user, 'reset')

        # reply with correct answer
        update_body_with_challenge(req_body, resp_body)
        req_body['challengeAnswer'] = user['reset_code']
        resp = register_account(req_body)
        self.assertEqual(resp.status_code, requests.codes.ok)

        #TODO check that new login works
        #TODO check that messages are deleted

