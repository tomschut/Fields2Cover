#=============================================================================
#    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
#                     Author: Tom Schut
#                        BSD-3 License
#=============================================================================

import connexion
import fields2cover as f2c
from openapi_server.models.error import Error
from openapi_server.models.swaths import Swaths
from openapi_server.models.generate_swaths_request import GenerateSwathsRequest
from openapi_server.controllers.converters import swaths_to_model, model_to_cells


def generate_swaths(body):
    """Generate swaths
    
    Based on: f2c::sg::BruteForce bf;
              F2CSwaths swaths = bf.generateSwaths(M_PI, robot.getCovWidth(), no_hl.getGeometry(0));
    """
    if connexion.request.is_json:
        generate_swaths_request = GenerateSwathsRequest.from_dict(connexion.request.get_json())
        angle = generate_swaths_request.angle
        width = generate_swaths_request.width
        cells_model = generate_swaths_request.cells
        
        if angle is None or width is None or cells_model is None:
            return Error(code='MISSING_PARAMS', message='angle, width, and cells required'), 400
        
        try:
            cells = model_to_cells(cells_model)
            
            # Use BruteForce swath generator
            swath_gen = f2c.SG_BruteForce()
            
            # Generate swaths on first cell
            if cells.size() > 0:
                swaths = swath_gen.generateSwaths(float(angle), float(width), cells.getGeometry(0))
            else:
                return Error(code='EMPTY_CELLS', message='No cells provided'), 400
            
            return swaths_to_model(swaths), 200
        except Exception as e:
            return Error(code='SWATH_ERROR', message=str(e)), 400
    
    return Error(code='INVALID_REQUEST', message='Request must be JSON'), 400
