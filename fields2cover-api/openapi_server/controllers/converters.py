#=============================================================================
#    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
#                     Author: Tom Schut
#                        BSD-3 License
#=============================================================================
# Conversion utilities between OpenAPI models and Fields2Cover objects

import fields2cover as f2c
from openapi_server.models.model_field import ModelField
from openapi_server.models.cells import Cells as CellsModel
from openapi_server.models.cell import Cell as CellModel
from openapi_server.models.linear_ring import LinearRing as LinearRingModel
from openapi_server.models.point import Point as PointModel
from openapi_server.models.robot import Robot as RobotModel
from openapi_server.models.swaths import Swaths as SwathsModel
from openapi_server.models.swath import Swath as SwathModel
from openapi_server.models.line_string import LineString as LineStringModel
from openapi_server.models.path import Path as PathModel
from openapi_server.models.path_state import PathState as PathStateModel


# =============================================================================
# Point conversions
# =============================================================================

def point_to_dict(point):
    """Convert f2c Point to dict"""
    return {
        'x': float(point.getX()),
        'y': float(point.getY()),
        'z': float(point.getZ()) if hasattr(point, 'getZ') else 0.0
    }


def dict_to_point(d):
    """Convert dict to f2c Point"""
    return f2c.Point(float(d['x']), float(d['y']), float(d.get('z', 0.0)))


def model_to_point(model):
    """Convert PointModel to f2c Point"""
    return f2c.Point(float(model.x), float(model.y), float(model.z) if model.z else 0.0)


def point_to_model(point):
    """Convert f2c Point to PointModel"""
    return PointModel(
        x=float(point.getX()),
        y=float(point.getY()),
        z=float(point.getZ()) if hasattr(point, 'getZ') else 0.0
    )


# =============================================================================
# LinearRing conversions
# =============================================================================

def linearring_to_list(ring):
    """Convert f2c LinearRing to list of point dicts"""
    return [point_to_dict(ring.getGeometry(i)) for i in range(ring.size())]


def list_to_linearring(points):
    """Convert list of point dicts to f2c LinearRing"""
    ring = f2c.LinearRing()
    for p in points:
        ring.addPoint(dict_to_point(p))
    return ring


def model_to_linearring(model):
    """Convert LinearRingModel to f2c LinearRing"""
    ring = f2c.LinearRing()
    if model.points:
        for p in model.points:
            ring.addPoint(model_to_point(p))
    return ring


def linearring_to_model(ring):
    """Convert f2c LinearRing to LinearRingModel"""
    points = [point_to_model(ring.getGeometry(i)) for i in range(ring.size())]
    return LinearRingModel(points=points)


# =============================================================================
# Cell conversions
# =============================================================================

def cell_to_dict(cell):
    """Convert f2c Cell to dict"""
    rings = []
    for i in range(cell.size()):
        rings.append(linearring_to_list(cell.getGeometry(i)))
    return {'rings': rings}


def dict_to_cell(d):
    """Convert dict to f2c Cell"""
    if not d.get('rings') or len(d['rings']) == 0:
        return f2c.Cell()
    
    # First ring is outer ring
    outer_ring = list_to_linearring(d['rings'][0])
    cell = f2c.Cell(outer_ring)
    
    # Additional rings are holes
    for i in range(1, len(d['rings'])):
        inner_ring = list_to_linearring(d['rings'][i])
        cell.addRing(inner_ring)
    
    return cell


def model_to_cell(model):
    """Convert CellModel to f2c Cell"""
    if not model.outer_ring:
        return f2c.Cell()
    
    outer_ring = model_to_linearring(model.outer_ring)
    cell = f2c.Cell(outer_ring)
    
    if model.inner_rings:
        for inner_model in model.inner_rings:
            inner_ring = model_to_linearring(inner_model)
            cell.addRing(inner_ring)
    
    return cell


def cell_to_model(cell):
    """Convert f2c Cell to CellModel"""
    outer_ring = linearring_to_model(cell.getGeometry(0)) if cell.size() > 0 else None
    inner_rings = [linearring_to_model(cell.getGeometry(i)) for i in range(1, cell.size())]
    
    return CellModel(
        outer_ring=outer_ring,
        inner_rings=inner_rings if inner_rings else None,
        area=float(cell.area())
    )


# =============================================================================
# Cells conversions
# =============================================================================

def cells_to_dict(cells):
    """Convert f2c Cells to dict"""
    return {
        'cells': [cell_to_dict(cells.getGeometry(i)) for i in range(cells.size())]
    }


def dict_to_cells(d):
    """Convert dict to f2c Cells"""
    cells = f2c.Cells()
    if d.get('cells'):
        for cell_data in d['cells']:
            cells.addGeometry(dict_to_cell(cell_data))
    return cells


def model_to_cells(model):
    """Convert CellsModel to f2c Cells"""
    cells = f2c.Cells()
    if model.cells:
        for cell_model in model.cells:
            cells.addGeometry(model_to_cell(cell_model))
    return cells


def cells_to_model(cells):
    """Convert f2c Cells to CellsModel"""
    cell_models = [cell_to_model(cells.getGeometry(i)) for i in range(cells.size())]
    return CellsModel(cells=cell_models)


# =============================================================================
# Field conversions
# =============================================================================

def field_to_model(field):
    """Convert f2c Field to ModelField"""
    cells_model = cells_to_model(field.getField())
    
    return ModelField(
        id=field.getId() if field.getId() else None,
        crs=field.getCRS() if field.getCRS() else None,
        cells=cells_model,
        area=float(field.getField().area())
    )


def model_to_field(model):
    """Convert ModelField to f2c Field"""
    field = f2c.Field()
    if model.id:
        field.setId(model.id)
    if model.crs:
        field.setCRS(model.crs)
    
    if model.cells:
        field.setField(model_to_cells(model.cells))
    
    return field


# =============================================================================
# Robot conversions
# =============================================================================

def robot_to_model(robot):
    """Convert f2c Robot to RobotModel"""
    return RobotModel(
        robot_width=float(robot.getWidth()),
        coverage_width=float(robot.getCovWidth()),
        min_turning_radius=float(robot.getMinTurningRadius())
    )


def model_to_robot(model):
    """Convert RobotModel to f2c Robot"""
    robot = f2c.Robot(float(model.robot_width), float(model.coverage_width))
    if model.min_turning_radius is not None:
        robot.setMinTurningRadius(float(model.min_turning_radius))
    return robot


# =============================================================================
# LineString/Swath conversions
# =============================================================================

def linestring_to_model(linestring):
    """Convert f2c LineString to LineStringModel"""
    points = [point_to_model(linestring.getGeometry(i)) for i in range(linestring.size())]
    return LineStringModel(points=points)


def model_to_linestring(model):
    """Convert LineStringModel to f2c LineString"""
    linestring = f2c.LineString()
    if model.points:
        for p in model.points:
            linestring.addPoint(model_to_point(p))
    return linestring


def swath_to_model(swath):
    """Convert f2c Swath to SwathModel"""
    return SwathModel(
        path=linestring_to_model(swath.getPath()),
        width=float(swath.getWidth()),
        id=int(swath.getId())
    )


def model_to_swath(model):
    """Convert SwathModel to f2c Swath"""
    path = model_to_linestring(model.path)
    swath = f2c.Swath(path, float(model.width))
    if model.id is not None:
        swath.setId(int(model.id))
    return swath


def swaths_to_model(swaths):
    """Convert f2c Swaths to SwathsModel"""
    # f2c Swaths is a vector-like container, use .size() method
    swath_models = []
    for i in range(swaths.size()):
        swath_models.append(swath_to_model(swaths[i]))
    return SwathsModel(swaths=swath_models)


def model_to_swaths(model):
    """Convert SwathsModel to f2c Swaths"""
    swaths = f2c.Swaths()
    if model.swaths:
        for swath_model in model.swaths:
            # Extract the LineString path and width from the swath
            path = model_to_linestring(swath_model.path)
            width = float(swath_model.width)
            # Use the correct append signature: append(LineString, width)
            swaths.append(path, width)
    return swaths


# =============================================================================
# Path conversions
# =============================================================================

def pathstate_to_model(state):
    """Convert f2c PathState to PathStateModel"""
    # Map f2c PathSectionType enum values to strings
    type_map = {
        0: 'SWATH',
        1: 'TURN', 
        2: 'DRIVE'
    }
    
    # Get the type value and convert to string
    type_value = int(state.type) if hasattr(state, 'type') else 1
    type_str = type_map.get(type_value, 'TURN')
    
    return PathStateModel(
        point=point_to_model(state.point),
        angle=float(state.angle),
        velocity=float(state.velocity),
        length=float(state.len),
        type=type_str,
        curvature=float(state.curvature) if hasattr(state, 'curvature') else None
    )


def path_to_model(path):
    """Convert f2c Path to PathModel"""
    states = []
    # f2c Path uses .size() method
    for i in range(path.size()):
        states.append(pathstate_to_model(path[i]))
    return PathModel(
        states=states,
        length=float(path.length())
    )


def model_to_pathstate(model):
    """Convert PathStateModel to f2c PathState"""
    # Map string types back to f2c enum values
    type_map = {
        'SWATH': 0,
        'TURN': 1,
        'DRIVE': 2
    }
    
    point = model_to_point(model.point)
    
    # Create PathState with just the point
    state = f2c.PathState()
    state.point = point
    state.angle = float(model.angle)
    state.velocity = float(model.velocity)
    
    if model.length is not None:
        state.len = float(model.length)
    if model.type:
        state.type = type_map.get(model.type, 1)  # Default to TURN
    if model.curvature is not None:
        state.curvature = float(model.curvature)
    
    return state


def model_to_path(model):
    """Convert PathModel to f2c Path"""
    # f2c Path is constructed from a list/vector of states
    # We need to create states and then build the path
    
    states_list = []
    if model.states:
        for state_model in model.states:
            # Create PathState
            state = f2c.PathState()
            state.point = model_to_point(state_model.point)
            state.angle = float(state_model.angle)
            state.velocity = float(state_model.velocity)
            state.len = float(state_model.length) if state_model.length else 0.0
            
            # Map type back to enum
            type_map = {
                'SWATH': 0,
                'TURN': 1,
                'DRIVE': 2
            }
            state.type = type_map.get(state_model.type, 1)
            
            if state_model.curvature is not None:
                state.curvature = float(state_model.curvature)
            
            states_list.append(state)
    
    # Create path from states vector
    # f2c.Path constructor might accept a vector/list of states
    try:
        path = f2c.Path(states_list)
    except:
        # If that doesn't work, try to create empty path and populate it differently
        # The Path might be immutable or need special construction
        # For now, create a minimal path that f2c.Transform can work with
        path = f2c.Path()
        # Copy attributes if possible
        if hasattr(path, 'states'):
            path.states = states_list
    
    return path


# =============================================================================
# Route conversions
# =============================================================================

def route_to_model(route):
    """Convert f2c Route to model dict
    
    A Route is essentially a collection of swaths with connections
    """
    # Extract swaths from the route
    swaths_list = []
    for i in range(route.sizeVSwaths()):
        swath = route.getVSwaths(i)
        swaths_list.append(swath_to_model(swath))
    
    # Extract connection paths between swaths
    connections_list = []
    for i in range(route.sizeVConnections()):
        connection = route.getVConnection(i)
        # Connection is a Path object
        connections_list.append(path_to_model(connection))
    
    return {
        'swaths': swaths_list,
        'connections': connections_list
    }


def model_to_route(model):
    """Convert model dict to f2c Route"""
    route = f2c.Route()
    
    # Add swaths
    if 'swaths' in model:
        for swath_data in model['swaths']:
            if isinstance(swath_data, dict):
                swath = model_to_swath(SwathModel.from_dict(swath_data))
            else:
                swath = model_to_swath(swath_data)
            route.addSwath(swath)
    
    # Add connections if present
    if 'connections' in model and model['connections']:
        for connection_data in model['connections']:
            if isinstance(connection_data, dict):
                connection = model_to_path(PathModel.from_dict(connection_data))
            else:
                connection = model_to_path(connection_data)
            route.addConnection(connection)
    
    return route
