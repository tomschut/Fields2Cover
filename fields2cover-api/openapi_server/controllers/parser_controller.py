#=============================================================================
#    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
#                     Author: Tom Schut
#                        BSD-3 License
#=============================================================================

import connexion
import fields2cover as f2c
import os
from openapi_server.models.error import Error
from openapi_server.models.model_field import ModelField
from openapi_server.models.import_field_gml_request import ImportFieldGmlRequest
from openapi_server.controllers.converters import field_to_model


def import_field_gml(body):
    """Import field from GML file
    
    Based on: F2CField field = f2c::Parser::importFieldGml(...)
    """
    if connexion.request.is_json:
        import_field_gml_request = ImportFieldGmlRequest.from_dict(connexion.request.get_json())
        file_path = import_field_gml_request.file_path
        
        if not file_path:
            return Error(code='MISSING_PATH', message='filePath is required'), 400
        
        if not os.path.exists(file_path):
            return Error(code='FILE_NOT_FOUND', message=f'File not found: {file_path}'), 404
        
        try:
            parser = f2c.Parser()
            field = parser.importFieldGml(file_path)
            return field_to_model(field), 200
        except Exception as e:
            return Error(code='IMPORT_ERROR', message=str(e)), 400
    
    return Error(code='INVALID_REQUEST', message='Request must be JSON'), 400
