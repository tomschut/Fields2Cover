import unittest

from flask import json

from openapi_server.models.error import Error  # noqa: E501
from openapi_server.models.model_field import ModelField  # noqa: E501
from openapi_server.test import BaseTestCase


class TestFieldController(BaseTestCase):
    """FieldController integration test stubs"""

    def test_clone_field(self):
        """Test case for clone_field

        Clone a field
        """
        model_field = openapi_server.ModelField()
        headers = { 
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        }
        response = self.client.open(
            '/api/v1/field/clone',
            method='POST',
            headers=headers,
            data=json.dumps(model_field),
            content_type='application/json')
        self.assert200(response,
                       'Response body is : ' + response.data.decode('utf-8'))


if __name__ == '__main__':
    unittest.main()
