#=============================================================================
#    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
#                     Author: Tom Schut
#                        BSD-3 License
#=============================================================================
# Integration test suite for Fields2Cover API

import unittest
import requests
import json
import math
import os
import time
from pathlib import Path


class Fields2CoverAPIIntegrationTests(unittest.TestCase):
    """Integration tests for Fields2Cover API endpoints"""
    
    BASE_URL = "http://localhost:8080/api/v1"
    TEST_DATA_PATH = None
    
    @classmethod
    def setUpClass(cls):
        """Set up test fixtures"""
        # Find the test data path
        current_dir = Path(__file__).parent.parent.parent
        cls.TEST_DATA_PATH = str(current_dir / "data" / "test1.xml")
        
        # Check if server is running
        try:
            response = requests.get(f"{cls.BASE_URL}/ui/", timeout=5)
            print(f"\n✅ Server is running at {cls.BASE_URL}")
            print(f"📚 Swagger UI available at: {cls.BASE_URL}/ui/")
        except requests.exceptions.ConnectionError:
            raise unittest.SkipTest(
                f"⚠️  Server is not running at {cls.BASE_URL}. "
                "Please start the server with: python -m openapi_server"
            )
        except requests.exceptions.Timeout:
            raise unittest.SkipTest(
                f"⚠️  Server at {cls.BASE_URL} is not responding. "
                "Please check if the server is running."
            )
    
    def setUp(self):
        """Set up before each test"""
        self.session = requests.Session()
        self.session.headers.update({'Content-Type': 'application/json'})
    
    def tearDown(self):
        """Clean up after each test"""
        self.session.close()
    
    # =========================================================================
    # Test 1: Parser Controller - Import Field
    # =========================================================================
    
    def test_01_import_field_gml(self):
        """Test importing a field from GML file"""
        print("\n🧪 Testing: Import Field GML")
        
        payload = {
            "filePath": self.TEST_DATA_PATH
        }
        
        response = self.session.post(
            f"{self.BASE_URL}/parser/import-field-gml",
            json=payload
        )
        
        self.assertEqual(response.status_code, 200, 
                        f"Expected 200, got {response.status_code}: {response.text}")
        
        data = response.json()
        self.assertIn('cells', data)
        self.assertIn('area', data)
        self.assertGreater(data['area'], 0)
        
        print(f"   ✅ Field imported successfully with area: {data['area']:.2f} m²")
        
        return data
    
    def test_02_import_field_not_found(self):
        """Test importing non-existent file"""
        print("\n🧪 Testing: Import Field - File Not Found")
        
        payload = {
            "filePath": "/nonexistent/file.xml"
        }
        
        response = self.session.post(
            f"{self.BASE_URL}/parser/import-field-gml",
            json=payload
        )
        
        self.assertEqual(response.status_code, 404)
        print("   ✅ Correctly returned 404 for non-existent file")
    
    # =========================================================================
    # Test 2: Field Controller - Clone
    # =========================================================================
    
    def test_03_clone_field(self):
        """Test cloning a field"""
        print("\n🧪 Testing: Clone Field")
        
        # First import a field
        field_data = self.test_01_import_field_gml()
        
        response = self.session.post(
            f"{self.BASE_URL}/field/clone",
            json=field_data
        )
        
        self.assertEqual(response.status_code, 200, 
                        f"Expected 200, got {response.status_code}: {response.text}")
        
        cloned_data = response.json()
        self.assertAlmostEqual(cloned_data['area'], field_data['area'], places=2)
        
        print(f"   ✅ Field cloned successfully")
    
    # =========================================================================
    # Test 3: Transform Controller - UTM Conversion
    # =========================================================================
    
    def test_04_transform_to_utm(self):
        """Test transforming field to UTM coordinates"""
        print("\n🧪 Testing: Transform to UTM")
        
        # Import field first
        field_data = self.test_01_import_field_gml()
        
        payload = {
            "field": field_data,
            "isETRS89": True
        }
        
        response = self.session.post(
            f"{self.BASE_URL}/transform/to-utm",
            json=payload
        )
        
        if response.status_code != 200:
            print(f"   ❌ Transform failed: {response.status_code}")
            print(f"   Response: {response.text}")
        
        self.assertEqual(response.status_code, 200, 
                        f"Transform failed: {response.text}")
        
        utm_field = response.json()
        self.assertIn('cells', utm_field)
        
        print("   ✅ Field transformed to UTM successfully")
        
        return utm_field
    
    # =========================================================================
    # Test 4: Robot Controller - Create Robot
    # =========================================================================
    
    def test_05_create_robot(self):
        """Test creating a robot configuration"""
        print("\n🧪 Testing: Create Robot")
        
        payload = {
            "robotWidth": 2.0,
            "coverageWidth": 6.0,
            "minTurningRadius": 2.0
        }
        
        response = self.session.post(
            f"{self.BASE_URL}/robot/create",
            json=payload
        )
        
        self.assertEqual(response.status_code, 200, 
                        f"Expected 200, got {response.status_code}: {response.text}")
        
        robot = response.json()
        self.assertAlmostEqual(robot['robotWidth'], 2.0, places=5)
        self.assertAlmostEqual(robot['coverageWidth'], 6.0, places=5)
        self.assertAlmostEqual(robot['minTurningRadius'], 2.0, places=5)
        
        print(f"   ✅ Robot created: width={robot['robotWidth']}m, "
              f"coverage={robot['coverageWidth']}m")
        
        return robot
    
    def test_06_create_robot_invalid_width(self):
        """Test creating robot with invalid parameters"""
        print("\n🧪 Testing: Create Robot - Invalid Width")
        
        payload = {
            "robotWidth": -1.0,
            "coverageWidth": 6.0
        }
        
        response = self.session.post(
            f"{self.BASE_URL}/robot/create",
            json=payload
        )
        
        self.assertEqual(response.status_code, 400)
        print("   ✅ Correctly rejected negative width")
    
    # =========================================================================
    # Test 5: Headland Controller - Generate Headlands
    # =========================================================================
    
    def test_07_generate_headlands(self):
        """Test generating headlands"""
        print("\n🧪 Testing: Generate Headlands")
        
        # Get UTM field
        utm_field = self.test_04_transform_to_utm()
        
        payload = {
            "cells": utm_field['cells'],
            "width": 18.0  # 3.0 * 6.0 (robot width)
        }
        
        response = self.session.post(
            f"{self.BASE_URL}/headland/generate",
            json=payload
        )
        
        if response.status_code != 200:
            print(f"   ❌ Headland generation failed: {response.status_code}")
            print(f"   Response: {response.text}")
        
        self.assertEqual(response.status_code, 200,
                        f"Headland generation failed: {response.text}")
        
        headlands = response.json()
        self.assertIn('cells', headlands)
        
        print(f"   ✅ Headlands generated successfully")
        
        return headlands
    
    # =========================================================================
    # Test 6: Swath Controller - Generate Swaths
    # =========================================================================
    
    def test_08_generate_swaths(self):
        """Test generating swaths"""
        print("\n🧪 Testing: Generate Swaths")
        
        # Get headlands
        headlands = self.test_07_generate_headlands()
        
        payload = {
            "angle": math.pi,
            "width": 6.0,
            "cells": headlands
        }
        
        response = self.session.post(
            f"{self.BASE_URL}/swath/generate",
            json=payload
        )
        
        if response.status_code != 200:
            print(f"   ❌ Swath generation failed: {response.status_code}")
            print(f"   Response: {response.text}")
        
        self.assertEqual(response.status_code, 200,
                        f"Swath generation failed: {response.text}")
        
        swaths = response.json()
        self.assertIn('swaths', swaths)
        self.assertGreater(len(swaths['swaths']), 0)
        
        print(f"   ✅ Generated {len(swaths['swaths'])} swaths")
        
        return swaths
    
    # =========================================================================
    # Test 7: Route Controller - Sort Swaths
    # =========================================================================
    
    def test_09_sort_swaths(self):
        """Test sorting swaths with snake pattern"""
        print("\n🧪 Testing: Sort Swaths")
        
        # Get swaths
        swaths = self.test_08_generate_swaths()
        
        payload = {
            "swaths": swaths,
            "variant": 0
        }
        
        response = self.session.post(
            f"{self.BASE_URL}/route/sort-swaths",
            json=payload
        )
        
        if response.status_code != 200:
            print(f"   ❌ Swath sorting failed: {response.status_code}")
            print(f"   Response: {response.text}")
        
        self.assertEqual(response.status_code, 200,
                        f"Swath sorting failed: {response.text}")
        
        sorted_swaths = response.json()
        self.assertIn('swaths', sorted_swaths)
        self.assertEqual(len(sorted_swaths['swaths']), len(swaths['swaths']))
        
        print(f"   ✅ Swaths sorted successfully ({len(sorted_swaths['swaths'])} swaths)")
        
        return sorted_swaths
    
    def test_10_sort_swaths_invalid_variant(self):
        """Test sorting with invalid variant"""
        print("\n🧪 Testing: Sort Swaths - Invalid Variant")
        
        swaths = self.test_08_generate_swaths()
        
        payload = {
            "swaths": swaths,
            "variant": 10  # Invalid variant
        }
        
        response = self.session.post(
            f"{self.BASE_URL}/route/sort-swaths",
            json=payload
        )
        
        self.assertEqual(response.status_code, 400)
        print("   ✅ Correctly rejected invalid variant")
    
    # =========================================================================
    # Test 8: Path Controller - Plan Path
    # =========================================================================
    
    def test_11_plan_path_dubins(self):
        """Test path planning with Dubins curves"""
        print("\n🧪 Testing: Plan Path - Dubins")
        
        robot = self.test_05_create_robot()
        sorted_swaths = self.test_09_sort_swaths()
        
        payload = {
            "robot": robot,
            "swaths": sorted_swaths,
            "turningAlgorithm": "DUBINS"
        }
        
        response = self.session.post(
            f"{self.BASE_URL}/path/plan",
            json=payload
        )
        
        if response.status_code != 200:
            print(f"   ❌ Path planning failed: {response.status_code}")
            print(f"   Response: {response.text}")
        
        self.assertEqual(response.status_code, 200,
                        f"Path planning failed: {response.text}")
        
        path = response.json()
        self.assertIn('states', path)
        self.assertIn('length', path)
        self.assertGreater(len(path['states']), 0)
        self.assertGreater(path['length'], 0)
        
        print(f"   ✅ Path planned: {len(path['states'])} states, "
              f"length={path['length']:.2f}m")
        
        return path
    
    def test_12_plan_path_reeds_shepp(self):
        """Test path planning with Reeds-Shepp curves"""
        print("\n🧪 Testing: Plan Path - Reeds-Shepp")
        
        robot = self.test_05_create_robot()
        sorted_swaths = self.test_09_sort_swaths()
        
        payload = {
            "robot": robot,
            "swaths": sorted_swaths,
            "turningAlgorithm": "REEDS_SHEPP"
        }
        
        response = self.session.post(
            f"{self.BASE_URL}/path/plan",
            json=payload
        )
        
        if response.status_code != 200:
            print(f"   ❌ Path planning failed: {response.status_code}")
            print(f"   Response: {response.text}")
        
        self.assertEqual(response.status_code, 200,
                        f"Path planning failed: {response.text}")
        
        path = response.json()
        self.assertGreater(len(path['states']), 0)
        
        print(f"   ✅ Reeds-Shepp path planned: {len(path['states'])} states")
    
    def test_13_plan_path_invalid_algorithm(self):
        """Test path planning with invalid algorithm"""
        print("\n🧪 Testing: Plan Path - Invalid Algorithm")
        
        robot = self.test_05_create_robot()
        sorted_swaths = self.test_09_sort_swaths()
        
        payload = {
            "robot": robot,
            "swaths": sorted_swaths,
            "turningAlgorithm": "INVALID_ALGO"
        }
        
        response = self.session.post(
            f"{self.BASE_URL}/path/plan",
            json=payload
        )
        
        self.assertEqual(response.status_code, 400)
        print("   ✅ Correctly rejected invalid algorithm")
    
    # # =========================================================================
    # # Test 9: Transform Controller - Transform Back to Previous CRS
    # # =========================================================================
    
    # def test_14_transform_to_prev_crs(self):
    #     """Test transforming path back to original CRS"""
    #     print("\n🧪 Testing: Transform to Previous CRS")
        
    #     # Get path and original field
    #     path = self.test_11_plan_path_dubins()
    #     utm_field = self.test_04_transform_to_utm()
        
    #     payload = {
    #         "path": path,
    #         "field": utm_field
    #     }
        
    #     response = self.session.post(
    #         f"{self.BASE_URL}/transform/to-prev-crs",
    #         json=payload
    #     )
        
    #     # Expect 501 Not Implemented
    #     self.assertEqual(response.status_code, 501,
    #                     f"Expected 501 Not Implemented, got {response.status_code}: {response.text}")
        
    #     print(f"   ⚠️  Transform to prev CRS: Not yet implemented (expected)")
        
    #     return None

    # =========================================================================
    # Test 9: Route Controller - Sort Swaths with Different Algorithms
    # =========================================================================
    
    def test_14_sort_swaths_boustrophedon_variants(self):
        """Test sorting swaths with different Boustrophedon variants"""
        print("\n🧪 Testing: Sort Swaths - Boustrophedon Variants")
        
        swaths = self.test_08_generate_swaths()
        
        for variant in range(4):  # Test variants 0-3
            print(f"   Testing variant {variant}...")
            payload = {
                "swaths": swaths,
                "algorithm": "BOUSTROPHEDON",
                "variant": variant
            }
            
            response = self.session.post(
                f"{self.BASE_URL}/route/sort-swaths",
                json=payload
            )
            
            self.assertEqual(response.status_code, 200,
                            f"Variant {variant} failed: {response.text}")
            
            sorted_swaths = response.json()
            self.assertEqual(len(sorted_swaths['swaths']), len(swaths['swaths']))
        
        print(f"   ✅ All 4 Boustrophedon variants tested successfully")
    
    def test_15_sort_swaths_snake(self):
        """Test sorting swaths with SNAKE algorithm (alias for Boustrophedon)"""
        print("\n🧪 Testing: Sort Swaths - Snake Pattern")
        
        swaths = self.test_08_generate_swaths()
        
        payload = {
            "swaths": swaths,
            "algorithm": "SNAKE",
            "variant": 0
        }
        
        response = self.session.post(
            f"{self.BASE_URL}/route/sort-swaths",
            json=payload
        )
        
        self.assertEqual(response.status_code, 200,
                        f"Snake sorting failed: {response.text}")
        
        sorted_swaths = response.json()
        self.assertIn('swaths', sorted_swaths)
        self.assertEqual(len(sorted_swaths['swaths']), len(swaths['swaths']))
        
        print(f"   ✅ Snake pattern sorted successfully ({len(sorted_swaths['swaths'])} swaths)")
    
    def test_16_sort_swaths_spiral(self):
        """Test sorting swaths with SPIRAL algorithm"""
        print("\n🧪 Testing: Sort Swaths - Spiral Pattern")
        
        swaths = self.test_08_generate_swaths()
        
        payload = {
            "swaths": swaths,
            "algorithm": "SPIRAL"
        }
        
        response = self.session.post(
            f"{self.BASE_URL}/route/sort-swaths",
            json=payload
        )
        
        self.assertEqual(response.status_code, 200,
                        f"Spiral sorting failed: {response.text}")
        
        sorted_swaths = response.json()
        self.assertIn('swaths', sorted_swaths)
        self.assertEqual(len(sorted_swaths['swaths']), len(swaths['swaths']))
        
        print(f"   ✅ Spiral pattern sorted successfully ({len(sorted_swaths['swaths'])} swaths)")
        
        return sorted_swaths
    
    def test_17_sort_swaths_invalid_algorithm(self):
        """Test sorting with invalid algorithm"""
        print("\n🧪 Testing: Sort Swaths - Invalid Algorithm")
        
        swaths = self.test_08_generate_swaths()
        
        payload = {
            "swaths": swaths,
            "algorithm": "INVALID_ALGORITHM"
        }
        
        response = self.session.post(
            f"{self.BASE_URL}/route/sort-swaths",
            json=payload
        )
        
        self.assertEqual(response.status_code, 400)
        error = response.json()
        self.assertIn('INVALID_ALGORITHM', error.get('code', ''))
        
        print("   ✅ Correctly rejected invalid algorithm")
    
    def test_18_compare_sorting_algorithms(self):
        """Test and compare different sorting algorithms"""
        print("\n🧪 Testing: Compare Sorting Algorithms")
        
        swaths = self.test_08_generate_swaths()
        
        algorithms = [
            ("BOUSTROPHEDON", 0),
            ("SNAKE", 0),
            ("SPIRAL", None)
        ]
        
        results = {}
        
        for algo, variant in algorithms:
            payload = {"swaths": swaths, "algorithm": algo}
            if variant is not None:
                payload["variant"] = variant
            
            response = self.session.post(
                f"{self.BASE_URL}/route/sort-swaths",
                json=payload
            )
            
            self.assertEqual(response.status_code, 200,
                            f"{algo} failed: {response.text}")
            
            sorted_swaths = response.json()
            results[algo] = len(sorted_swaths['swaths'])
            print(f"      {algo}: {results[algo]} swaths")
        
        print("   ✅ All sorting algorithms compared successfully")
        
        # All should have same number of swaths, just different order
        self.assertEqual(len(set(results.values())), 1,
                        "All algorithms should return same number of swaths")

    # =========================================================================
    # Test 10: Complete Flow Integration Test
    # =========================================================================
    
    def test_19_complete_flow_integration(self):
        """Test complete workflow from field import to path planning"""
        print("\n🧪 Testing: Complete Integration Flow")
        print("   This replicates the 8_complete_flow.py tutorial")
        
        # Step 1: Import field
        print("   1️⃣  Importing field...")
        field_response = self.session.post(
            f"{self.BASE_URL}/parser/import-field-gml",
            json={"filePath": self.TEST_DATA_PATH}
        )
        self.assertEqual(field_response.status_code, 200,
                        f"Import failed: {field_response.text}")
        field = field_response.json()
        field_area = field['area']
        
        # Step 2: Transform to UTM
        print("   2️⃣  Transforming to UTM...")
        utm_response = self.session.post(
            f"{self.BASE_URL}/transform/to-utm",
            json={"field": field, "isETRS89": True}
        )
        
        if utm_response.status_code != 200:
            print(f"   ❌ UTM transform failed: {utm_response.status_code}")
            print(f"   Response: {utm_response.text}")
        
        self.assertEqual(utm_response.status_code, 200,
                        f"UTM transform failed: {utm_response.text}")
        utm_field = utm_response.json()
        
        # Step 3: Create robot
        print("   3️⃣  Creating robot configuration...")
        robot_response = self.session.post(
            f"{self.BASE_URL}/robot/create",
            json={
                "robotWidth": 2.0,
                "coverageWidth": 6.0,
                "minTurningRadius": 2.0
            }
        )
        self.assertEqual(robot_response.status_code, 200,
                        f"Robot creation failed: {robot_response.text}")
        robot = robot_response.json()
        
        # Step 4: Generate headlands
        print("   4️⃣  Generating headlands...")
        headland_response = self.session.post(
            f"{self.BASE_URL}/headland/generate",
            json={
                "cells": utm_field['cells'],
                "width": 3.0 * robot['robotWidth']
            }
        )
        self.assertEqual(headland_response.status_code, 200,
                        f"Headland generation failed: {headland_response.text}")
        no_hl = headland_response.json()
        
        # Step 5: Generate swaths
        print("   5️⃣  Generating swaths...")
        swath_response = self.session.post(
            f"{self.BASE_URL}/swath/generate",
            json={
                "angle": math.pi,
                "width": robot['coverageWidth'],
                "cells": no_hl
            }
        )
        self.assertEqual(swath_response.status_code, 200,
                        f"Swath generation failed: {swath_response.text}")
        swaths = swath_response.json()
        
        # Step 6: Sort swaths with different algorithms
        print("   6️⃣  Sorting swaths with Spiral algorithm...")
        sort_response = self.session.post(
            f"{self.BASE_URL}/route/sort-swaths",
            json={"swaths": swaths, "algorithm": "SPIRAL"}
        )
        self.assertEqual(sort_response.status_code, 200,
                        f"Swath sorting failed: {sort_response.text}")
        sorted_swaths = sort_response.json()
        
        # Step 7: Plan path
        print("   7️⃣  Planning path with Dubins curves...")
        path_response = self.session.post(
            f"{self.BASE_URL}/path/plan",
            json={
                "robot": robot,
                "swaths": sorted_swaths,
                "turningAlgorithm": "DUBINS"
            }
        )
        
        if path_response.status_code != 200:
            print(f"   ❌ Path planning failed: {path_response.status_code}")
            print(f"   Response: {path_response.text}")
        
        self.assertEqual(path_response.status_code, 200,
                        f"Path planning failed: {path_response.text}")
        path = path_response.json()
        
        print("\n   ✅ Complete flow executed successfully!")
        print(f"      - Field area: {field_area:.2f} m²")
        print(f"      - Swaths generated: {len(sorted_swaths['swaths'])}")
        print(f"      - Path states: {len(path['states'])}")
        print(f"      - Path length: {path['length']:.2f} m")



# =============================================================================
# Test Runner
# =============================================================================

def run_tests():
    """Run all integration tests"""
    # Create test suite
    loader = unittest.TestLoader()
    suite = loader.loadTestsFromTestCase(Fields2CoverAPIIntegrationTests)
    
    # Run tests with verbose output
    runner = unittest.TextTestRunner(verbosity=2)
    result = runner.run(suite)
    
    # Print summary
    print("\n" + "=" * 70)
    print("TEST SUMMARY")
    print("=" * 70)
    print(f"Tests run: {result.testsRun}")
    print(f"Successes: {result.testsRun - len(result.failures) - len(result.errors)}")
    print(f"Failures: {len(result.failures)}")
    print(f"Errors: {len(result.errors)}")
    print("=" * 70)
    
    return result.wasSuccessful()


if __name__ == '__main__':
    success = run_tests()
    exit(0 if success else 1)
