import unittest

from flask import json

from openapi_server.models.error import Error  # noqa: E501
from openapi_server.models.generate_swaths_request import GenerateSwathsRequest  # noqa: E501
from openapi_server.models.swaths import Swaths  # noqa: E501
from openapi_server.test import BaseTestCase


class TestSwathController(BaseTestCase):
    """SwathController integration test stubs"""

    def test_generate_swaths(self):
        """Test case for generate_swaths

        Generate swaths
        """
        generate_swaths_request = openapi_server.GenerateSwathsRequest()
        headers = { 
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        }
        response = self.client.open(
            '/api/v1/swath/generate',
            method='POST',
            headers=headers,
            data=json.dumps(generate_swaths_request),
            content_type='application/json')
        self.assert200(response,
                       'Response body is : ' + response.data.decode('utf-8'))


if __name__ == '__main__':
    unittest.main()
