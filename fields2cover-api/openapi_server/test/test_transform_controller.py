import unittest

from flask import json

from openapi_server.models.error import Error  # noqa: E501
from openapi_server.models.model_field import ModelField  # noqa: E501
from openapi_server.models.path import Path  # noqa: E501
from openapi_server.models.transform_to_prev_crs_request import TransformToPrevCRSRequest  # noqa: E501
from openapi_server.models.transform_to_utm_request import TransformToUTMRequest  # noqa: E501
from openapi_server.test import BaseTestCase


class TestTransformController(BaseTestCase):
    """TransformController integration test stubs"""

    def test_transform_to_prev_crs(self):
        """Test case for transform_to_prev_crs

        Transform to previous CRS
        """
        transform_to_prev_crs_request = openapi_server.TransformToPrevCRSRequest()
        headers = { 
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        }
        response = self.client.open(
            '/api/v1/transform/to-previous-crs',
            method='POST',
            headers=headers,
            data=json.dumps(transform_to_prev_crs_request),
            content_type='application/json')
        self.assert200(response,
                       'Response body is : ' + response.data.decode('utf-8'))

    def test_transform_to_utm(self):
        """Test case for transform_to_utm

        Transform field to UTM coordinates
        """
        transform_to_utm_request = openapi_server.TransformToUTMRequest()
        headers = { 
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        }
        response = self.client.open(
            '/api/v1/transform/to-utm',
            method='POST',
            headers=headers,
            data=json.dumps(transform_to_utm_request),
            content_type='application/json')
        self.assert200(response,
                       'Response body is : ' + response.data.decode('utf-8'))


if __name__ == '__main__':
    unittest.main()
