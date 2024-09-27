from datetime import date, datetime  # noqa: F401

from typing import List, Dict  # noqa: F401

from openapi_server.models.base_model import Model
from openapi_server.models.geometry_element import GeometryElement
from openapi_server import util

from openapi_server.models.geometry_element import GeometryElement  # noqa: E501

class Polygon(Model):
    """NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).

    Do not edit the class manually.
    """

    def __init__(self, type=None, bbox=None, coordinates=None):  # noqa: E501
        """Polygon - a model defined in OpenAPI

        :param type: The type of this Polygon.  # noqa: E501
        :type type: str
        :param bbox: The bbox of this Polygon.  # noqa: E501
        :type bbox: List[float]
        :param coordinates: The coordinates of this Polygon.  # noqa: E501
        :type coordinates: List[List[List[float]]]
        """
        self.openapi_types = {
            'type': str,
            'bbox': List[float],
            'coordinates': List[List[List[float]]]
        }

        self.attribute_map = {
            'type': 'type',
            'bbox': 'bbox',
            'coordinates': 'coordinates'
        }

        self._type = type
        self._bbox = bbox
        self._coordinates = coordinates

    @classmethod
    def from_dict(cls, dikt) -> 'Polygon':
        """Returns the dict as a model

        :param dikt: A dict.
        :type: dict
        :return: The Polygon of this Polygon.  # noqa: E501
        :rtype: Polygon
        """
        return util.deserialize_model(dikt, cls)

    @property
    def type(self) -> str:
        """Gets the type of this Polygon.


        :return: The type of this Polygon.
        :rtype: str
        """
        return self._type

    @type.setter
    def type(self, type: str):
        """Sets the type of this Polygon.


        :param type: The type of this Polygon.
        :type type: str
        """
        allowed_values = ["Point", "MultiPoint", "LineString", "MultiLineString", "Polygon", "MultiPolygon"]  # noqa: E501
        if type not in allowed_values:
            raise ValueError(
                "Invalid value for `type` ({0}), must be one of {1}"
                .format(type, allowed_values)
            )

        self._type = type

    @property
    def bbox(self) -> List[float]:
        """Gets the bbox of this Polygon.

        A GeoJSON object MAY have a member named \"bbox\" to include information on the coordinate range for its Geometries, Features, or FeatureCollections. The value of the bbox member MUST be an array of length 2*n where n is the number of dimensions represented in the contained geometries, with all axes of the most southwesterly point followed by all axes of the more northeasterly point. The axes order of a bbox follows the axes order of geometries.   # noqa: E501

        :return: The bbox of this Polygon.
        :rtype: List[float]
        """
        return self._bbox

    @bbox.setter
    def bbox(self, bbox: List[float]):
        """Sets the bbox of this Polygon.

        A GeoJSON object MAY have a member named \"bbox\" to include information on the coordinate range for its Geometries, Features, or FeatureCollections. The value of the bbox member MUST be an array of length 2*n where n is the number of dimensions represented in the contained geometries, with all axes of the most southwesterly point followed by all axes of the more northeasterly point. The axes order of a bbox follows the axes order of geometries.   # noqa: E501

        :param bbox: The bbox of this Polygon.
        :type bbox: List[float]
        """

        self._bbox = bbox

    @property
    def coordinates(self) -> List[List[List[float]]]:
        """Gets the coordinates of this Polygon.


        :return: The coordinates of this Polygon.
        :rtype: List[List[List[float]]]
        """
        return self._coordinates

    @coordinates.setter
    def coordinates(self, coordinates: List[List[List[float]]]):
        """Sets the coordinates of this Polygon.


        :param coordinates: The coordinates of this Polygon.
        :type coordinates: List[List[List[float]]]
        """
        if coordinates is None:
            raise ValueError("Invalid value for `coordinates`, must not be `None`")  # noqa: E501

        self._coordinates = coordinates
