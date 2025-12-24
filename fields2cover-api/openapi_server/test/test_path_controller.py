import unittest

from flask import json

from openapi_server.models.error import Error  # noqa: E501
from openapi_server.models.path import Path  # noqa: E501
from openapi_server.models.plan_path_request import PlanPathRequest  # noqa: E501
from openapi_server.test import BaseTestCase


class TestPathController(BaseTestCase):
    """PathController integration test stubs"""

    def test_plan_path(self):
        """Test case for plan_path

        Plan coverage path
        """
        plan_path_request = openapi_server.PlanPathRequest()
        headers = { 
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        }
        response = self.client.open(
            '/api/v1/path/plan',
            method='POST',
            headers=headers,
            data=json.dumps(plan_path_request),
            content_type='application/json')
        self.assert200(response,
                       'Response body is : ' + response.data.decode('utf-8'))


if __name__ == '__main__':
    unittest.main()
