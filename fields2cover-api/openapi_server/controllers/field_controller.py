#=============================================================================
#    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
#                     Author: Tom Schut
#                        BSD-3 License
#=============================================================================

import connexion
import fields2cover as f2c
from openapi_server.models.error import Error
from openapi_server.models.model_field import ModelField
from openapi_server.controllers.converters import field_to_model, model_to_field


def clone_field(body):
    """Clone a field
    
    Based on: F2CField orig_field = field.clone();
    """
    if connexion.request.is_json:
        model_field = ModelField.from_dict(connexion.request.get_json())
        
        try:
            field = model_to_field(model_field)
            cloned_field = field.clone()
            return field_to_model(cloned_field), 200
        except Exception as e:
            return Error(code='CLONE_ERROR', message=str(e)), 400
    
    return Error(code='INVALID_REQUEST', message='Request must be JSON'), 400
