#=============================================================================
#    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
#                     Author: Tom Schut
#                        BSD-3 License
#=============================================================================

import connexion
import fields2cover as f2c
from openapi_server.models.error import Error
from openapi_server.models.robot import Robot
from openapi_server.models.create_robot_request import CreateRobotRequest
from openapi_server.controllers.converters import robot_to_model


def create_robot(body):
    """Create a robot configuration
    
    Based on: F2CRobot robot (2.0, 6.0);
              robot.setMinTurningRadius(2);
    """
    if connexion.request.is_json:
        create_robot_request = CreateRobotRequest.from_dict(connexion.request.get_json())
        robot_width = create_robot_request.robot_width
        coverage_width = create_robot_request.coverage_width
        min_turning_radius = create_robot_request.min_turning_radius if create_robot_request.min_turning_radius is not None else 0.0
        
        if robot_width is None or coverage_width is None:
            return Error(code='MISSING_PARAMS', message='robotWidth and coverageWidth required'), 400
        
        if robot_width <= 0 or coverage_width <= 0:
            return Error(code='INVALID_PARAMS', message='Width values must be positive'), 400
        
        try:
            robot = f2c.Robot(float(robot_width), float(coverage_width))
            robot.setMinTurningRadius(float(min_turning_radius))
            return robot_to_model(robot), 200
        except Exception as e:
            return Error(code='CREATE_ERROR', message=str(e)), 400
    
    return Error(code='INVALID_REQUEST', message='Request must be JSON'), 400
