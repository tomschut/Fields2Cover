#!/bin/bash
#=============================================================================
#    Copyright (C) 2021-2024 Wageningen University - All Rights Reserved
#                     Author: Tom Schut
#                        BSD-3 License
#=============================================================================

set -e

echo "🔧 Installing Fields2Cover to virtual environment..."

# Activate venv
cd /home/tom/devenv/Fields2Cover/
source .venv/bin/activate

# Get venv paths
VENV_PATH=$(pwd)/.venv
PYTHON_VERSION=$(python -c "import sys; print(f'python{sys.version_info.major}.{sys.version_info.minor}')")

echo "📍 Virtual environment: $VENV_PATH"
echo "🐍 Python version: $PYTHON_VERSION"

# Go to Fields2Cover root
cd /home/tom/devenv/Fields2Cover

# Clean previous builds
echo "🧹 Cleaning previous builds..."
rm -rf build
mkdir build && cd build

# Configure with venv installation paths
echo "⚙️  Configuring CMake..."
cmake -DCMAKE_BUILD_TYPE=Release \
      -DBUILD_PYTHON=ON \
      -DBUILD_TESTS=OFF \
      -DBUILD_TUTORIALS=OFF \
      -DCMAKE_INSTALL_PREFIX="$VENV_PATH" \
      -DPYTHON_EXECUTABLE="$VENV_PATH/bin/python" \
      -DCMAKE_INSTALL_RPATH="$VENV_PATH/lib" \
      -DCMAKE_INSTALL_RPATH_USE_LINK_PATH=TRUE \
      ..

# Build
echo "🔨 Building..."
make -j$(nproc)

# Install to venv
echo "📦 Installing to venv..."
make install

# The Python module might be in a different location, let's find and symlink it
echo "🔗 Setting up Python module..."
PYTHON_MODULE=$(find . -name "_fields2cover*.so" | head -n 1)
if [ -n "$PYTHON_MODULE" ]; then
    SITE_PACKAGES="$VENV_PATH/lib/$PYTHON_VERSION/site-packages"
    mkdir -p "$SITE_PACKAGES"
    
    # Copy Python module files
    cp -v swig/python/fields2cover.py "$SITE_PACKAGES/" 2>/dev/null || true
    cp -v swig/python/_fields2cover*.so "$SITE_PACKAGES/" 2>/dev/null || true
    cp -v python/fields2cover.py "$SITE_PACKAGES/" 2>/dev/null || true
    cp -v python/_fields2cover*.so "$SITE_PACKAGES/" 2>/dev/null || true
    
    echo "✅ Python module installed to: $SITE_PACKAGES"
fi

# Find where OR-Tools is installed
echo "🔍 Looking for OR-Tools library..."
ORTOOLS_LIB=$(find /usr/local/lib* /usr/lib* ~/.local/lib* "$VENV_PATH/lib" -name "libortools.so*" 2>/dev/null | head -n 1)

if [ -n "$ORTOOLS_LIB" ]; then
    ORTOOLS_DIR=$(dirname "$ORTOOLS_LIB")
    echo "✅ Found OR-Tools at: $ORTOOLS_DIR"
else
    echo "⚠️  OR-Tools not found in standard locations"
    ORTOOLS_DIR="/usr/local/lib"
fi

# Create a wrapper script that sets LD_LIBRARY_PATH
cat > "$VENV_PATH/bin/activate-f2c" << EOF
#!/bin/bash
# Activate venv with Fields2Cover and all dependencies

# Source the regular activation
source "$VENV_PATH/bin/activate"

# Add library paths
export LD_LIBRARY_PATH="$VENV_PATH/lib:$ORTOOLS_DIR:\$LD_LIBRARY_PATH"

echo "✅ Virtual environment activated with Fields2Cover"
echo "📚 Library path includes: $VENV_PATH/lib and $ORTOOLS_DIR"
EOF

chmod +x "$VENV_PATH/bin/activate-f2c"

# Also create a sitecustomize.py to set the path automatically
cat > "$SITE_PACKAGES/sitecustomize.py" << EOF
# Auto-configure library paths for Fields2Cover
import os
import sys

venv_lib = "$VENV_PATH/lib"
ortools_lib = "$ORTOOLS_DIR"

if 'LD_LIBRARY_PATH' in os.environ:
    os.environ['LD_LIBRARY_PATH'] = f"{venv_lib}:{ortools_lib}:" + os.environ['LD_LIBRARY_PATH']
else:
    os.environ['LD_LIBRARY_PATH'] = f"{venv_lib}:{ortools_lib}"
EOF

# Update library path for the session
export LD_LIBRARY_PATH="$VENV_PATH/lib:$ORTOOLS_DIR:$LD_LIBRARY_PATH"

# Verify installation
echo ""
echo "✅ Verifying installation..."
cd /home/tom/devenv/Fields2Cover
python -c "import fields2cover as f2c; print('✅ fields2cover imported successfully!')" || {
    echo "❌ Import still failed. Debugging..."
    echo ""
    echo "Current LD_LIBRARY_PATH:"
    echo "$LD_LIBRARY_PATH"
    echo ""
    echo "Python path:"
    python -c "import sys; print('\n'.join(sys.path))"
    echo ""
    echo "Looking for fields2cover files:"
    find "$VENV_PATH" -name "*fields2cover*" 2>/dev/null
    echo ""
    echo "Looking for libortools:"
    find /usr /home -name "libortools.so*" 2>/dev/null | head -5
    echo ""
    echo "Checking ldd on the Python module:"
    MODULE=$(find "$SITE_PACKAGES" -name "_fields2cover*.so" | head -n 1)
    if [ -n "$MODULE" ]; then
        ldd "$MODULE" | grep -i "not found" || echo "All dependencies found"
    fi
    exit 1
}

echo ""
echo "✨ Installation complete!"
echo "📍 Libraries installed to: $VENV_PATH/lib"
echo "📍 Headers installed to: $VENV_PATH/include"
echo "🐍 Python module installed to: $VENV_PATH/lib/$PYTHON_VERSION/site-packages"
echo ""
echo "⚠️  IMPORTANT: To use Fields2Cover, always use the custom activation:"
echo "  source .venv/bin/activate-f2c"
echo ""
echo "Or set the library path manually:"
echo "  export LD_LIBRARY_PATH=\"$VENV_PATH/lib:$ORTOOLS_DIR:\$LD_LIBRARY_PATH\""
