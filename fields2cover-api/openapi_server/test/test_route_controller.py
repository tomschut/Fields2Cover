import unittest

from flask import json

from openapi_server.models.error import Error  # noqa: E501
from openapi_server.models.sort_swaths_request import SortSwathsRequest  # noqa: E501
from openapi_server.models.swaths import Swaths  # noqa: E501
from openapi_server.test import BaseTestCase


class TestRouteController(BaseTestCase):
    """RouteController integration test stubs"""

    def test_sort_swaths(self):
        """Test case for sort_swaths

        Sort swaths
        """
        sort_swaths_request = openapi_server.SortSwathsRequest()
        headers = { 
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        }
        response = self.client.open(
            '/api/v1/route/sort-swaths',
            method='POST',
            headers=headers,
            data=json.dumps(sort_swaths_request),
            content_type='application/json')
        self.assert200(response,
                       'Response body is : ' + response.data.decode('utf-8'))


if __name__ == '__main__':
    unittest.main()
