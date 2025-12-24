import unittest

from flask import json

from openapi_server.models.error import Error  # noqa: E501
from openapi_server.models.save_visualization200_response import SaveVisualization200Response  # noqa: E501
from openapi_server.models.save_visualization_request import SaveVisualizationRequest  # noqa: E501
from openapi_server.test import BaseTestCase


class TestVisualizerController(BaseTestCase):
    """VisualizerController integration test stubs"""

    def test_save_visualization(self):
        """Test case for save_visualization

        Save visualization
        """
        save_visualization_request = openapi_server.SaveVisualizationRequest()
        headers = { 
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        }
        response = self.client.open(
            '/api/v1/visualizer/save',
            method='POST',
            headers=headers,
            data=json.dumps(save_visualization_request),
            content_type='application/json')
        self.assert200(response,
                       'Response body is : ' + response.data.decode('utf-8'))


if __name__ == '__main__':
    unittest.main()
