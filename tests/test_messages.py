# pylint: disable=missing-docstring

import base64
import json
import requests
from requests_toolbelt.multipart.encoder import MultipartEncoder

from . import base
from . import settings

def b64e(plain):
    return base64.b64encode(plain)

def b64d(encoded):
    return base64.b64decode(encoded)

class MessagesTest(base.BaseTest):
    user = settings.EXISTING_USERS[1]

    def get_list(self, query_params=None):
        query_params = query_params or {}
        return requests.get(
            self.url_prefix(self.user) + '/messages/',
            params=query_params,
            **self.auth_good())

    def create_message_json(self, data, auth=None):
        auth = auth or {}
        return requests.post(
            self.url_prefix(self.user) + '/messages/',
            headers={'content-type': 'application/json'},
            data=json.dumps(data),
            **auth)

    def create_message_multipart(self, encoder, auth=None):
        auth = auth or {}
        return requests.post(
            self.url_prefix(self.user) + '/messages/',
            headers={'content-type': encoder.content_type},
            data=encoder,
            **auth)

    def get_message(self, message_id):
        return requests.get(
            self.url_prefix(self.user) + '/messages/' + str(message_id),
            **self.auth_good())

    def get_attachments(self, message_id):
        return requests.get(
            self.url_prefix(self.user) + '/messages/' + str(message_id) +
            '/attachments',
            **self.auth_good())

    def modify_meta(self, message_id, last_modified, data):
        return requests.patch(
            self.url_prefix(self.user) + '/messages/' + str(message_id),
            params={'lastModified': last_modified},
            headers={'content-type': 'application/json'},
            data=json.dumps(data),
            **self.auth_good())

    def delete_message(self, message_id, last_modified):
        return requests.delete(
            self.url_prefix(self.user) + '/messages/' + str(message_id),
            params={'lastModified': last_modified},
            **self.auth_good())


    def subtest_get_list(self, messages):
        resp = self.get_list()
        self.assertEqual(resp.status_code, requests.codes.ok)
        json_result = json.loads(resp.text)
        self.assertEqual(json_result['resultsTotal'], len(messages))
        self.assertEqual(json_result['resultsReturned'], len(messages))
        self.assertEqual(len(json_result['data']), len(messages))
        for msg_remote, msg_local in zip(json_result['data'], messages):
            self.assertEqualOrAssign(msg_remote, 'id', msg_local)
            self.assertEqualOrAssign(msg_remote, 'lastModified', msg_local)

    def subtest_get_list_with_content(self, messages):
        resp = self.get_list({'includeData': True})
        self.assertEqual(resp.status_code, requests.codes.ok)
        remote_messages = json.loads(resp.text)['data']
        self.assertEqual(len(remote_messages), len(messages))
        for msg_remote, msg_local in zip(remote_messages, messages):
            self.assertEqual(msg_remote['id'], msg_local['id'])
            self.assertEqual(msg_remote['lastModified'], msg_local['lastModified'])
            self.assertTimestampIsNow(msg_remote['lastModified'])
            self.assertEqualOrAssign(msg_remote, 'dateReceived', msg_local)
            self.assertIsoTimeIsNow(msg_remote['dateReceived'])
            self.assertEqual(msg_remote['deleted'], msg_local['deleted'])
            self.assertEqual(b64d(msg_remote['meta']), msg_local['meta'])
            self.assertEqual(b64d(msg_remote['keySafe']), msg_local['keySafe'])
            self.assertEqual(b64d(msg_remote['content']), msg_local['content'])
            self.assertEqual(msg_remote['hasAttachments'], msg_local['hasAttachments'])

    def subtest_post_message_unauth(self):
        message = {
            'keySafe': 'I am the key safe',
            'content': 'I am a message',
            'meta': '',
            'hasAttachments': False,
            'deleted': False,
        }
        resp = self.create_message_json({
            'keySafe': b64e(message['keySafe']),
            'content': b64e(message['content']),
        })
        self.assertEqual(resp.status_code, requests.codes.ok)
        self.assertEqual(json.loads(resp.text), {})
        message['id'] = base.VALUE_NOT_AVAILABLE
        message['lastModified'] = base.VALUE_NOT_AVAILABLE
        message['dateReceived'] = base.VALUE_NOT_AVAILABLE
        return message

    def subtest_post_message_auth(self):
        message = {
            'keySafe': 'I am another key safe',
            'content': 'I am another message',
            'meta': 'I am meta',
            'attachments': 'I am an attachment'*1000,
            'hasAttachments': True,
            'deleted': False,
        }
        resp = self.create_message_json({
            'keySafe': b64e(message['keySafe']),
            'content': b64e(message['content']),
            'meta': b64e(message['meta']),
            'attachments': b64e(message['attachments']),
        }, self.auth_good())
        self.assertEqual(resp.status_code, requests.codes.ok)
        json_result = json.loads(resp.text)
        self.assertTrue(json_result.has_key('id'))
        self.assertTrue(json_result.has_key('lastModified'))
        self.assertIsoTimeIsNow(json_result['dateReceived'])
        message['id'] = json_result['id']
        message['lastModified'] = json_result['lastModified']
        message['dateReceived'] = json_result['dateReceived']
        return message

    def subtest_post_message_unauth_multipart(self):
        message = {
            'keySafe': 'Yet another key safe',
            'content': 'Yet another message',
            'meta': '',
            'attachments': 'Yet another attachment',
            'hasAttachments': True,
            'deleted': False,
        }
        encoder = MultipartEncoder({
            'keySafe': message['keySafe'],
            'content': message['content'],
            'attachments': message['attachments'],
        })
        resp = self.create_message_multipart(encoder)
        self.assertEqual(resp.status_code, requests.codes.ok)
        self.assertEqual(json.loads(resp.text), {})
        message['id'] = base.VALUE_NOT_AVAILABLE
        message['lastModified'] = base.VALUE_NOT_AVAILABLE
        message['dateReceived'] = base.VALUE_NOT_AVAILABLE
        return message

    def subtest_post_message_auth_multipart(self):
        message = {
            'keySafe': 'Yet another key safe',
            'content': 'Yet another message',
            'meta': 'Yet another meta',
            'hasAttachments': False,
            'deleted': False,
        }
        encoder = MultipartEncoder({
            'keySafe': message['keySafe'],
            'content': message['content'],
            'meta': message['meta'],
        })
        resp = self.create_message_multipart(encoder, self.auth_good())
        self.assertEqual(resp.status_code, requests.codes.ok)
        json_result = json.loads(resp.text)
        self.assertTrue(json_result.has_key('id'))
        self.assertTrue(json_result.has_key('lastModified'))
        self.assertIsoTimeIsNow(json_result['dateReceived'])
        message['id'] = json_result['id']
        message['lastModified'] = json_result['lastModified']
        message['dateReceived'] = json_result['dateReceived']
        return message


    def test_size_limits(self):
        template = {
            'keySafe': 'k',
            'content': 'c',
            'meta': '',
            'attachments': '',
        }
        messages = []

        # create keySafe empty
        msg = dict(template)
        msg['keySafe'] = ''
        messages.append(msg)

        # create keySafe too big
        msg = dict(template)
        msg['keySafe'] = 'k' * (1024 + 1)
        messages.append(msg)

        # create content empty
        msg = dict(template)
        msg['content'] = ''
        messages.append(msg)

        # create content too big
        msg = dict(template)
        msg['content'] = 'c' * (128 * 1024 + 1)
        messages.append(msg)

        # create meta too big
        msg = dict(template)
        msg['meta'] = 'm' * (1024 + 1)
        messages.append(msg)

        messages_json = list(messages)
        messages_multipart = list(messages)

        # create attachments too big (json)
        msg = dict(template)
        msg['attachments'] = 'a' * (16 * 1024 * 1024 + 1)
        messages_json.append(msg)

        # create attachments too big (multipart)
        msg = dict(template)
        msg['attachments'] = 'a' * (100 * 1024 * 1024 + 1)
        messages_multipart.append(msg)

        for msg in messages_json:
            resp = self.create_message_json({
                'keySafe': b64e(msg['keySafe']),
                'content': b64e(msg['content']),
                'meta': b64e(msg['meta']),
                'attachments': b64e(msg['attachments']),
            }, self.auth_good())
            self.assertEqual(resp.status_code, requests.codes.bad_request)

        for msg in messages_multipart:
            encoder = MultipartEncoder(msg)
            resp = self.create_message_multipart(encoder, self.auth_good())
            self.assertEqual(resp.status_code, requests.codes.bad_request)

        # modify meta too big
        resp = self.modify_meta(42, 23, {'meta': b64e('m' * (1024 + 1))})
        self.assertEqual(resp.status_code, requests.codes.bad_request)

    def test_messages(self):
        messages = []

        # get empty list
        self.subtest_get_list(messages)

        # send messages
        messages.append(self.subtest_post_message_unauth())
        messages.append(self.subtest_post_message_auth())
        messages.append(self.subtest_post_message_unauth_multipart())
        messages.append(self.subtest_post_message_auth_multipart())

        # get nonempty list
        self.subtest_get_list(messages)

        # get list with includeData
        self.subtest_get_list_with_content(messages)

        # get nonexistant attachments of first message
        resp = self.get_attachments(messages[0]['id'])
        self.assertEqual(resp.status_code, requests.codes.not_found)

        # get attachments of second message
        resp = self.get_attachments(messages[1]['id'])
        self.assertEqual(resp.status_code, requests.codes.ok)
        self.assertEqual(resp.headers['content-type'], 'application/octet-stream')
        self.assertEqual(resp.headers['content-length'], str(len(resp.text)))
        messages[1]['attachments'] = resp.text

        # get nonexistant message
        resp = self.get_message(42)
        self.assertEqual(resp.status_code, requests.codes.not_found)

        # modify meta of nonexistant message
        resp = self.modify_meta(42, 23, {'meta': 'foo'})
        self.assertEqual(resp.status_code, requests.codes.not_found)

        # modify first message meta with bad last_modified
        resp = self.modify_meta(messages[0]['id'], 23, {'meta': 'foo'})
        self.assertEqual(resp.status_code, requests.codes.conflict)

        # modify first message meta
        messages[0]['meta'] = 'I am the message meta'
        resp = self.modify_meta(
            messages[0]['id'], messages[0]['lastModified'],
            {'meta': b64e(messages[0]['meta'])})
        self.assertEqual(resp.status_code, requests.codes.ok)
        json_result = json.loads(resp.text)
        self.assertEqual(messages[0]['id'], json_result['id'])
        self.assertTrue(messages[0]['lastModified'] < json_result['lastModified'])
        messages[0]['lastModified'] = json_result['lastModified']

        # move messages[0] to back of list because its lastModified has been updated
        messages = messages[1:] + messages[0:1]

        # get message, verify modification
        resp = self.get_message(messages[-1]['id'])
        self.assertEqual(resp.status_code, requests.codes.ok)
        json_result = json.loads(resp.text)
        self.assertEqual(b64d(json_result['meta']), messages[-1]['meta'])
        self.assertEqual(json_result['deleted'], messages[-1]['deleted'])

        # get list, verify that only first (now last) has been modified
        self.subtest_get_list(messages)

        # get list with modified_after
        resp = self.get_list({'modifiedAfter': messages[-2]['lastModified']})
        self.assertEqual(resp.status_code, requests.codes.ok)
        self.assertEqual(json.loads(resp.text)['data'],
                         [self.extract_id_lastmodified(messages[-1])])

        # try to delete message with bad last_modified
        resp = self.delete_message(messages[0]['id'], 23)
        self.assertEqual(resp.status_code, requests.codes.conflict)

        # delete message
        resp = self.delete_message(messages[0]['id'], messages[0]['lastModified'])
        self.assertEqual(resp.status_code, requests.codes.ok)
        json_result = json.loads(resp.text)
        self.assertEqual(json_result['id'], messages[0]['id'])
        self.assertGreater(json_result['lastModified'], messages[0]['lastModified'])
        messages[0]['lastModified'] = json_result['lastModified']

        # move messages[0] to back of list because its lastModified has been updated
        messages = messages[1:] + messages[0:1]

        # get message, verify deleted
        resp = self.get_message(messages[-1]['id'])
        self.assertEqual(resp.status_code, requests.codes.ok)
        json_result = json.loads(resp.text)
        self.assertEqual(json_result['deleted'], True)
        self.assertEqual(json_result['dateReceived'], '')
        self.assertEqual(json_result['meta'], '')
        self.assertEqual(json_result['keySafe'], '')
        self.assertEqual(json_result['content'], '')
        self.assertEqual(json_result['hasAttachments'], False)

        # get attachments of deleted message
        resp = self.get_attachments(messages[-1]['id'])
        self.assertEqual(resp.status_code, requests.codes.not_found)

        # get list, verify modifications
        self.subtest_get_list(messages)
