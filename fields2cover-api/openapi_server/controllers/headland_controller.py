#=============================================================================
#    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
#                     Author: Tom Schut
#                        BSD-3 License
#=============================================================================

import connexion
import fields2cover as f2c
from openapi_server.models.error import Error
from openapi_server.models.cells import Cells
from openapi_server.models.generate_headlands_request import GenerateHeadlandsRequest
from openapi_server.controllers.converters import cells_to_model, model_to_cells


def generate_headlands(body):
    """Generate headlands
    
    Based on: f2c::hg::ConstHL const_hl;
              F2CCells no_hl = const_hl.generateHeadlands(field.getField(), 3.0 * robot.getWidth());
    """
    if connexion.request.is_json:
        generate_headlands_request = GenerateHeadlandsRequest.from_dict(connexion.request.get_json())
        cells_model = generate_headlands_request.cells
        width = generate_headlands_request.width
        
        if cells_model is None or width is None:
            return Error(code='MISSING_PARAMS', message='cells and width required'), 400
        
        try:
            cells = model_to_cells(cells_model)
            headland_gen = f2c.HG_Const_gen()
            no_headlands = headland_gen.generateHeadlands(cells, float(width))
            return cells_to_model(no_headlands), 200
        except Exception as e:
            return Error(code='HEADLAND_ERROR', message=str(e)), 400
    
    return Error(code='INVALID_REQUEST', message='Request must be JSON'), 400
