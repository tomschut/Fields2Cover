#=============================================================================
#    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
#                     Author: Tom Schut
#                        BSD-3 License
#=============================================================================

import connexion
import fields2cover as f2c
from openapi_server.models.error import Error
from openapi_server.models.path import Path
from openapi_server.models.plan_path_request import PlanPathRequest
from openapi_server.controllers.converters import (
    path_to_model, model_to_robot, model_to_swaths
)


def plan_path(body):
    """Plan coverage path
    
    Based on: f2c::pp::PathPlanning path_planner;
              f2c::pp::DubinsCurves dubins;
              F2CPath path = path_planner.planPath(robot, swaths, dubins);
    """
    if connexion.request.is_json:
        plan_path_request = PlanPathRequest.from_dict(connexion.request.get_json())
        robot_model = plan_path_request.robot
        swaths_model = plan_path_request.swaths
        turning_algorithm = plan_path_request.turning_algorithm
        
        if robot_model is None or swaths_model is None or turning_algorithm is None:
            return Error(code='MISSING_PARAMS', message='robot, swaths, and turningAlgorithm required'), 400
        
        try:
            robot = model_to_robot(robot_model)
            swaths = model_to_swaths(swaths_model)
            
            # Create path planner
            path_planner = f2c.PP_PathPlanning()
            
            # Select turning algorithm
            if turning_algorithm == 'DUBINS':
                turn_planner = f2c.PP_DubinsCurves()
            elif turning_algorithm == 'DUBINS_CC':
                turn_planner = f2c.PP_DubinsCurvesCC()
            elif turning_algorithm == 'REEDS_SHEPP':
                turn_planner = f2c.PP_ReedsSheppCurves()
            elif turning_algorithm == 'REEDS_SHEPP_HC':
                turn_planner = f2c.PP_ReedsSheppCurvesHC()
            else:
                return Error(code='INVALID_ALGORITHM', message=f'Unknown algorithm: {turning_algorithm}'), 400
            
            # Plan the path
            path = path_planner.planPath(robot, swaths, turn_planner)
            
            return path_to_model(path), 200
        except Exception as e:
            return Error(code='PATH_ERROR', message=str(e)), 400
    
    return Error(code='INVALID_REQUEST', message='Request must be JSON'), 400
