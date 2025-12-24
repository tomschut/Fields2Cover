import unittest

from flask import json

from openapi_server.models.error import Error  # noqa: E501
from openapi_server.models.import_field_gml_request import ImportFieldGmlRequest  # noqa: E501
from openapi_server.models.model_field import ModelField  # noqa: E501
from openapi_server.test import BaseTestCase


class TestParserController(BaseTestCase):
    """ParserController integration test stubs"""

    def test_import_field_gml(self):
        """Test case for import_field_gml

        Import field from GML file
        """
        import_field_gml_request = openapi_server.ImportFieldGmlRequest()
        headers = { 
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        }
        response = self.client.open(
            '/api/v1/parser/import-field-gml',
            method='POST',
            headers=headers,
            data=json.dumps(import_field_gml_request),
            content_type='application/json')
        self.assert200(response,
                       'Response body is : ' + response.data.decode('utf-8'))


if __name__ == '__main__':
    unittest.main()
