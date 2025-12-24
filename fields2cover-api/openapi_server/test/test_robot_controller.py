import unittest

from flask import json

from openapi_server.models.create_robot_request import CreateRobotRequest  # noqa: E501
from openapi_server.models.error import Error  # noqa: E501
from openapi_server.models.robot import Robot  # noqa: E501
from openapi_server.test import BaseTestCase


class TestRobotController(BaseTestCase):
    """RobotController integration test stubs"""

    def test_create_robot(self):
        """Test case for create_robot

        Create robot configuration
        """
        create_robot_request = openapi_server.CreateRobotRequest()
        headers = { 
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        }
        response = self.client.open(
            '/api/v1/robot/create',
            method='POST',
            headers=headers,
            data=json.dumps(create_robot_request),
            content_type='application/json')
        self.assert200(response,
                       'Response body is : ' + response.data.decode('utf-8'))


if __name__ == '__main__':
    unittest.main()
