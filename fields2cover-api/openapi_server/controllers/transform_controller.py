#=============================================================================
#    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
#                     Author: Tom Schut
#                        BSD-3 License
#=============================================================================

import connexion
import fields2cover as f2c
from openapi_server.models.error import Error
from openapi_server.controllers.converters import (
    field_to_model, model_to_field
)


def transform_to_utm(body):
    """Transform field to UTM coordinates
    
    Based on: f2c::Transform::transformToUTM(field);
    """
    if connexion.request.is_json:
        data = connexion.request.get_json()
        
        try:
            # Get field and isETRS89 from request
            from openapi_server.models.model_field import ModelField
            model_field = ModelField.from_dict(data.get('field'))
            is_etrs89 = data.get('isETRS89', True)
            
            # Convert to f2c Field
            field = model_to_field(model_field)
            
            # Transform to UTM
            f2c.Transform.transformToUTM(field, is_etrs89)
            
            # Convert back to model
            return field_to_model(field), 200
            
        except Exception as e:
            import traceback
            traceback.print_exc()
            return Error(code='TRANSFORM_ERROR', message=str(e)), 400
    
    return Error(code='INVALID_REQUEST', message='Request must be JSON'), 400


def transform_to_prev_crs(body):
    """Transform path back to previous CRS
    
    This endpoint is not yet implemented due to complexity of reconstructing
    f2c Path objects from API models and CRS tracking requirements.
    """
    return Error(
        code='NOT_IMPLEMENTED',
        message='Transform to previous CRS is not yet implemented. Use the UTM coordinates directly or implement client-side coordinate transformation.'
    ), 501
