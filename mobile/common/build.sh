#!/bin/bash

# P3 移动平台构建脚本
# 用于构建 Android 和 iOS 平台的共享库

set -e

# 检查 gomobile 是否已安装
if ! command -v gomobile &> /dev/null; then
    echo "gomobile 未安装，正在安装..."
    go install golang.org/x/mobile/cmd/gomobile@latest
    gomobile init
fi

# 检查参数
PLATFORM=$1
if [ -z "$PLATFORM" ]; then
    echo "未指定平台，将构建所有平台"
    PLATFORM="all"
fi

# 构建 Android 库
build_android() {
    echo "构建 Android 库..."
    mkdir -p ../android/app/libs
    gomobile bind -target=android -o ../android/app/libs/p3.aar -v ./
    echo "Android 库构建完成: ../android/app/libs/p3.aar"
}

# 构建 iOS 库
build_ios() {
    echo "构建 iOS 库..."
    mkdir -p ../ios/Frameworks
    gomobile bind -target=ios -o ../ios/Frameworks/P3.xcframework -v ./
    echo "iOS 库构建完成: ../ios/Frameworks/P3.xcframework"
}

# 根据平台构建
case $PLATFORM in
    "android")
        build_android
        ;;
    "ios")
        build_ios
        ;;
    "all")
        build_android
        build_ios
        ;;
    *)
        echo "未知平台: $PLATFORM"
        echo "支持的平台: android, ios, all"
        exit 1
        ;;
esac

echo "构建完成"
