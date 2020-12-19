# vim: set expandtab shiftwidth=4 :
# pylint: disable=missing-docstring

import base64
import hashlib
import psycopg2

from . import settings


def encode_login_key(login_key):
    return base64.b64encode(hashlib.sha512(login_key).digest())

def get_connection(connection_str):
    return psycopg2.connect(connection_str)

def add_user(cursor, user):
    cursor.execute(
        "INSERT INTO users (reset_code, plan_id) " +
        "VALUES (%s, (SELECT id FROM plans WHERE name = %s)) " +
        "RETURNING id",
        [user.get('reset_code', ''), user.get('plan', '')])
    user_id = cursor.fetchone()[0]
    cursor.execute(
        "INSERT INTO addresses (user_id, address) VALUES (%s, %s)",
        [user_id, user['address']])
    cursor.execute(
        "SELECT upsert_keys_symm(%s, %s, %s)",
        [user['address'], encode_login_key(user['loginKey']),
        user.get('privateDataKey', '')])
    if user.has_key('encryptionPrivkey'):
        cursor.execute(
            "INSERT INTO keys_asymm " +
            "(id, user_id, key_type, pubkey, privkey, valid_from, valid_until) " +
		    "VALUES " +
		    "(kullo_new_id('keys_asymm', %s), %s, %s, %s, %s, %s, %s)",
		    [user_id, user_id, 'enc',
		    user['encryptionPubkey'], user['encryptionPrivkey'],
		    '2000-01-01T00:00:00Z', '2222-01-01T00:00:00Z'])

def delete_user(cursor, user):
    cursor.execute(
        "DELETE FROM users u USING addresses a " +
        "WHERE u.id = a.user_id AND a.address = %s", [user['address']])

def setup():
    with get_connection(settings.DB_CONNECTION_STRING) as conn:
        with conn.cursor() as cursor:
            for usr in settings.USERS.itervalues():
                delete_user(cursor, usr)
            for usr in settings.EXISTING_USERS.itervalues():
                add_user(cursor, usr)
            for usr in settings.RESET_USERS.itervalues():
                add_user(cursor, usr)
