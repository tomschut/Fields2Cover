#=============================================================================
#    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
#                     Author: Tom Schut
#                        BSD-3 License
#=============================================================================

import connexion
import fields2cover as f2c
from openapi_server.models.error import Error
from openapi_server.models.swaths import Swaths
from openapi_server.models.sort_swaths_request import SortSwathsRequest
from openapi_server.controllers.converters import (
    swaths_to_model, model_to_swaths, 
    route_to_model, model_to_field
)


def sort_swaths(body):
    """Sort swaths using the specified sorting order
    
    Supports multiple algorithms:
    - BOUSTROPHEDON: Snake pattern (default)
    - SNAKE: Alias for Boustrophedon
    - SPIRAL: Spiral pattern from outside to inside
    - CUSTOM: Custom sorting (not implemented yet)
    
    Variant parameter controls the specific pattern (0-3 for Boustrophedon)
    """
    if connexion.request.is_json:
        data = connexion.request.get_json()
        
        try:
            # Get algorithm directly from JSON data
            algorithm = data.get('algorithm', 'BOUSTROPHEDON')            
            
            # Get swaths
            sort_swaths_request = SortSwathsRequest.from_dict(data)
            swaths = model_to_swaths(sort_swaths_request.swaths)
            
            # Get variant (default to 0 for snake pattern)
            variant = sort_swaths_request.variant if sort_swaths_request.variant is not None else 0
            
            # Clamp variant to valid range (0-3)
            variant = max(0, min(3, variant))
            
            # Validate and create appropriate route sorter based on algorithm
            algorithm_upper = algorithm.upper()
            
            if algorithm_upper in ['BOUSTROPHEDON', 'SNAKE']:
                route_sorter = f2c.RP_Boustrophedon()
                sorted_swaths = route_sorter.genSortedSwaths(swaths, variant)
            elif algorithm_upper == 'SPIRAL':
                route_sorter = f2c.RP_Spiral()
                sorted_swaths = route_sorter.genSortedSwaths(swaths)
            else:
                print(f"Debug: Rejecting invalid algorithm: {algorithm}")
                return Error(
                    code='INVALID_ALGORITHM',
                    message=f'Unknown sorting algorithm: {algorithm}. Use BOUSTROPHEDON, SNAKE, or SPIRAL'
                ), 400
            
            # Convert back to model
            return swaths_to_model(sorted_swaths), 200
            
        except Exception as e:
            import traceback
            traceback.print_exc()
            return Error(code='SORT_ERROR', message=str(e)), 400
    
    return Error(code='INVALID_REQUEST', message='Request must be JSON'), 400


def generate_route(body):
    """Generate route with turns between swaths
    
    Based on: F2CRoute route = route_planner.genRoute(field, swaths);
    This includes turn maneuvers between swaths.
    """
    if connexion.request.is_json:
        data = connexion.request.get_json()
        
        try:
            from openapi_server.models.model_field import ModelField
            
            # Get field and swaths from request
            field_model = ModelField.from_dict(data.get('field'))
            swaths_data = data.get('swaths')
            
            if not field_model or not swaths_data:
                return Error(code='MISSING_PARAMS', 
                           message='field and swaths required'), 400
            
            # Convert models to f2c objects
            field = model_to_field(field_model)
            swaths = model_to_swaths(swaths_data)
            
            # Create route planner and generate route
            route_planner = f2c.RP_RoutePlannerBase()
            route = route_planner.genRoute(field, swaths)
            
            # Convert back to model
            return route_to_model(route), 200
            
        except Exception as e:
            import traceback
            traceback.print_exc()
            return Error(code='ROUTE_ERROR', message=str(e)), 400
    
    return Error(code='INVALID_REQUEST', message='Request must be JSON'), 400
