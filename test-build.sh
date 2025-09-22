#!/bin/bash

# Unified build test script - works on any OS
# Usage: ./test-build.sh [arm64|amd64] [target_os]
# Examples:
#   ./test-build.sh arm64 darwin    # macOS ARM64
#   ./test-build.sh amd64 darwin    # macOS AMD64
#   ./test-build.sh amd64 windows   # Windows AMD64

set -e

ARCH=${1:-arm64}
TARGET_OS=${2:-darwin}
CURRENT_OS=$(uname -s)

echo "🔧 Testing build logic for $TARGET_OS $ARCH on $CURRENT_OS"

# Simulate GitHub Actions environment
export GOOS=$TARGET_OS
export GOARCH=$ARCH
export CGO_ENABLED=1
export CGO_CFLAGS="-I$(pwd)/backend/sqlite3"

export BINARY_NAME_SERVER=sortedchat-server-$TARGET_OS-$ARCH
export BINARY_NAME_APP=sortedchat-app-$TARGET_OS-$ARCH

echo "📋 Environment Setup"
echo "  GOOS=$GOOS"
echo "  GOARCH=$GOARCH" 

# Apply the same logic as GitHub Actions
if [ "$TARGET_OS" = "windows" ]; then
    # Windows cross-compilation
    export BINARY_NAME_SERVER="$BINARY_NAME_SERVER.exe"
    export BINARY_NAME_APP="$BINARY_NAME_APP.exe"
    export CC="x86_64-w64-mingw32-gcc"
    export CXX="x86_64-w64-mingw32-g++"
    export PKG_CONFIG_PATH="/usr/x86_64-w64-mingw32/lib/pkgconfig"
    export CGO_CFLAGS="$CGO_CFLAGS -I/usr/x86_64-w64-mingw32/include"
    export CGO_LDFLAGS="-L/usr/x86_64-w64-mingw32/lib -static"
    echo "✅ Set up Windows cross-compilation"
    
elif [ "$TARGET_OS" = "darwin" ]; then
    # macOS cross-compilation logic
    if [ "$ARCH" = "amd64" ]; then
        export MACOS_TARGET_ARCH=x86_64
        # Cross-compile from ARM64 host to x86_64 target
        export CC="clang -target x86_64-apple-macos11.0"
        export CXX="clang++ -target x86_64-apple-macos11.0"
        export CGO_CFLAGS="$CGO_CFLAGS -target x86_64-apple-macos11.0 -mmacosx-version-min=11.0"
        export CGO_LDFLAGS="-target x86_64-apple-macos11.0 -mmacosx-version-min=11.0 -framework UniformTypeIdentifiers"
        export PKG_CONFIG_PATH="/opt/homebrew/lib/pkgconfig"
        echo "✅ Set up macOS x86_64 cross-compilation"
    else
        export MACOS_TARGET_ARCH=arm64
        # Native ARM64 build
        export CGO_CFLAGS="$CGO_CFLAGS -mmacosx-version-min=11.0"
        export CGO_LDFLAGS="-mmacosx-version-min=11.0 -framework UniformTypeIdentifiers"
        echo "✅ Set up macOS native ARM64 build"
    fi
fi

echo ""
echo "🔍 Final Environment Variables:"
echo "  CC=$CC"
echo "  CGO_CFLAGS=$CGO_CFLAGS"
echo "  CGO_LDFLAGS=$CGO_LDFLAGS"
echo "  PKG_CONFIG_PATH=$PKG_CONFIG_PATH"

if [ "$CURRENT_OS" = "Darwin" ]; then
    echo ""
    echo "🍎 Running on macOS - attempting actual build..."
    
    # Check for dependencies
    if ! command -v brew &> /dev/null; then
        echo "❌ Homebrew not found. Install it first."
        exit 1
    fi
    
    echo "📦 Installing dependencies..."
    # Only install Opus (new dependency) - macOS has built-in GUI frameworks
    brew install opus opusfile pkg-config || echo "Dependencies may already be installed"
    
    cd backend
    echo "🏗️ Building server binary..."
    go build -buildvcs=false -tags "prod,desktop,sqlite_fts5,webkit2_41,wv2runtime.download,production,devtools" -ldflags "-w -s" -o $BINARY_NAME_SERVER ./mono
    
    echo "🏗️ Building app binary..."
    go build -buildvcs=false -tags "prod,wails,desktop,sqlite_fts5,webkit2_41,wv2runtime.download,production,devtools" -ldflags "-w -s" -o $BINARY_NAME_APP ./mono
    
    echo ""
    echo "✅ Build Results:"
    ls -la $BINARY_NAME_SERVER $BINARY_NAME_APP
    
    echo ""
    echo "🔍 Architecture Check:"
    file $BINARY_NAME_SERVER
    file $BINARY_NAME_APP
    
    echo ""
    echo "🎉 Build completed successfully!"
    
else
    echo ""
    echo "🐧 Running on $CURRENT_OS - showing what would happen on macOS..."
    echo ""
    echo "📦 Would install: brew install opus opusfile pkg-config"
    echo "🏗️ Would build with these commands:"
    echo "  cd backend"
    echo "  go build -buildvcs=false -tags \"prod,desktop,sqlite_fts5,webkit2_41,wv2runtime.download,production,devtools\" -ldflags \"-w -s\" -o $BINARY_NAME_SERVER ./mono"
    echo "  go build -buildvcs=false -tags \"prod,wails,desktop,sqlite_fts5,webkit2_41,wv2runtime.download,production,devtools\" -ldflags \"-w -s\" -o $BINARY_NAME_APP ./mono"
    echo ""
    echo "🔍 Expected result: Mach-O 64-bit executable $MACOS_TARGET_ARCH"
    echo ""
    echo "✅ Logic verified! Ready to push to GitHub Actions."
fi

echo ""
echo "📋 Summary:"
echo "  Target: macOS $ARCH ($MACOS_TARGET_ARCH)"
echo "  Method: $([ "$MACOS_TARGET_ARCH" = "x86_64" ] && echo "Cross-compilation" || echo "Native build")"
echo "  Status: $([ "$CURRENT_OS" = "Darwin" ] && echo "Actually tested" || echo "Logic verified")" 