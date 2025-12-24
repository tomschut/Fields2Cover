import unittest

from flask import json

from openapi_server.models.cells import Cells  # noqa: E501
from openapi_server.models.error import Error  # noqa: E501
from openapi_server.models.generate_headlands_request import GenerateHeadlandsRequest  # noqa: E501
from openapi_server.test import BaseTestCase


class TestHeadlandController(BaseTestCase):
    """HeadlandController integration test stubs"""

    def test_generate_headlands(self):
        """Test case for generate_headlands

        Generate headlands
        """
        generate_headlands_request = openapi_server.GenerateHeadlandsRequest()
        headers = { 
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        }
        response = self.client.open(
            '/api/v1/headland/generate',
            method='POST',
            headers=headers,
            data=json.dumps(generate_headlands_request),
            content_type='application/json')
        self.assert200(response,
                       'Response body is : ' + response.data.decode('utf-8'))


if __name__ == '__main__':
    unittest.main()
