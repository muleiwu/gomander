#!/bin/bash

set -e

cleanup() {
    if [ -f "myapp.pid" ]; then
        ./myapp stop >/dev/null 2>&1 || true
    fi
}

trap cleanup EXIT

echo "=== Gomander 功能测试 ==="
echo ""

# 编译
echo "1. 编译程序..."
go build -o myapp
echo "✓ 编译成功"
echo ""

# 测试自定义命令
echo "2. 测试自定义命令..."
VERSION_OUTPUT=$(./myapp version)
if [ "$VERSION_OUTPUT" = "gomander example v1.0.0" ]; then
    echo "✓ 自定义 version 命令执行成功"
else
    echo "✗ 自定义 version 命令输出异常: $VERSION_OUTPUT"
    exit 1
fi
echo ""

# 测试 daemon 模式
echo "3. 测试 daemon 模式启动..."
./myapp start -d
sleep 2
echo "✓ Daemon 启动成功"
echo ""

# 检查 PID 文件
echo "4. 检查 PID 文件..."
if [ -f "myapp.pid" ]; then
    PID=$(cat myapp.pid)
    echo "✓ PID 文件存在: PID=$PID"
else
    echo "✗ PID 文件不存在"
    exit 1
fi
echo ""

# 检查日志文件
echo "5. 检查日志文件..."
if [ -f "myapp.log" ]; then
    echo "✓ 日志文件存在"
    echo "最新日志:"
    tail -3 myapp.log | sed 's/^/  /'
else
    echo "✗ 日志文件不存在"
    exit 1
fi
echo ""

# 测试 stop 命令
echo "6. 测试 stop 命令..."
./myapp stop
sleep 1
echo "✓ Stop 命令执行成功"
echo ""

# 验证 PID 文件被删除
echo "7. 验证 PID 文件已删除..."
if [ ! -f "myapp.pid" ]; then
    echo "✓ PID 文件已删除"
else
    echo "✗ PID 文件仍然存在"
    exit 1
fi
echo ""

echo "=== 所有测试通过！ ==="
