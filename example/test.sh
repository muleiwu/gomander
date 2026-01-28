#!/bin/bash

set -e

echo "=== Gomander 功能测试 ==="
echo ""

# 编译
echo "1. 编译程序..."
go build -o myapp
echo "✓ 编译成功"
echo ""

# 测试 daemon 模式
echo "2. 测试 daemon 模式启动..."
./myapp -d
sleep 2
echo "✓ Daemon 启动成功"
echo ""

# 检查 PID 文件
echo "3. 检查 PID 文件..."
if [ -f "myapp.pid" ]; then
    PID=$(cat myapp.pid)
    echo "✓ PID 文件存在: PID=$PID"
else
    echo "✗ PID 文件不存在"
    exit 1
fi
echo ""

# 检查日志文件
echo "4. 检查日志文件..."
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
echo "5. 测试 stop 命令..."
./myapp stop
sleep 1
echo "✓ Stop 命令执行成功"
echo ""

# 验证 PID 文件被删除
echo "6. 验证 PID 文件已删除..."
if [ ! -f "myapp.pid" ]; then
    echo "✓ PID 文件已删除"
else
    echo "✗ PID 文件仍然存在"
    exit 1
fi
echo ""

echo "=== 所有测试通过！ ==="
