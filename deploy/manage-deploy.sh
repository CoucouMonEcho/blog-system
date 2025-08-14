#!/bin/bash

# Blog System 部署管理脚本
set -e

# 配置变量
DEPLOY_PATH="/opt/blog-system"

# 日志函数
log_info() {
    echo "[INFO] $1"
}

log_error() {
    echo "[ERROR] $1"
}

# 静默执行函数
silent_exec() {
    "$@" >/dev/null 2>&1
}

# 显示帮助信息
show_help() {
    echo "Blog System 部署管理脚本"
    echo ""
    echo "用法: $0 [选项] [服务名]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示此帮助信息"
    echo "  -l, --list     列出所有可用服务"
    echo "  -s, --status   显示所有服务状态"
    echo "  -r, --restart  重启指定服务"
    echo "  -t, --stop     停止指定服务"
    echo "  -u, --update   更新指定服务"
    echo ""
    echo "服务名:"
    echo "  all            所有服务"
    echo "  user           用户服务"
    echo "  content        内容服务"
    echo "  stat           统计服务"
    echo "  gateway        网关服务"
    echo "  common         公共模块"
    echo ""
    echo "示例:"
    echo "  $0 -s                    # 显示所有服务状态"
    echo "  $0 -u user              # 更新用户服务"
    echo "  $0 -r all               # 重启所有服务"
    echo "  $0 -t gateway           # 停止网关服务"
}

# 列出所有服务
list_services() {
    echo "可用服务:"
    echo "  - user-service (端口: 8001)"
    echo "  - gateway-service (端口: 8000)"
    echo "  - common (公共模块)"
    echo ""
    echo "服务状态:"
    systemctl status user-service --no-pager -l | head -5
    systemctl status content-service --no-pager -l | head -5
    systemctl status stat-service --no-pager -l | head -5
    systemctl status gateway-service --no-pager -l | head -5
}

# 显示服务状态
show_status() {
    local service=$1
    
    if [ "$service" = "all" ] || [ -z "$service" ]; then
        log_info "=== 所有服务状态 ==="
        systemctl status user-service --no-pager -l
        echo ""
        systemctl status content-service --no-pager -l
        echo ""
        systemctl status stat-service --no-pager -l
        echo ""
        systemctl status gateway-service --no-pager -l
        echo ""
        log_info "=== 端口监听状态 ==="
        netstat -tlnp | grep -E ":(8000|8001|8002|8003|8004)" || echo "未发现相关端口监听"
    else
        log_info "=== $service 服务状态 ==="
        systemctl status ${service}-service --no-pager -l
    fi
}

# 重启服务
restart_service() {
    local service=$1
    
    if [ "$service" = "all" ]; then
        log_info "重启所有服务..."
        silent_exec systemctl restart user-service
        silent_exec systemctl restart content-service
        silent_exec systemctl restart stat-service
        silent_exec systemctl restart gateway-service
        log_info "所有服务重启完成"
    else
        log_info "重启 $service 服务..."
        silent_exec systemctl restart ${service}-service
        log_info "$service 服务重启完成"
    fi
}

# 停止服务
stop_service() {
    local service=$1
    
    if [ "$service" = "all" ]; then
        log_info "停止所有服务..."
        silent_exec systemctl stop user-service
        silent_exec systemctl stop gateway-service
        log_info "所有服务已停止"
    else
        log_info "停止 $service 服务..."
        silent_exec systemctl stop ${service}-service
        log_info "$service 服务已停止"
    fi
}

# 更新服务
update_service() {
    local service=$1
    
    if [ "$service" = "all" ]; then
        log_info "更新所有服务..."
        cd ${DEPLOY_PATH}
        export BLOG_PASSWORD="$BLOG_PASSWORD"
        ./deploy/deploy.sh
    elif [ "$service" = "user" ]; then
        log_info "更新用户服务..."
        cd ${DEPLOY_PATH}
        export BLOG_PASSWORD="$BLOG_PASSWORD"
        ./deploy/deploy-user-service.sh
    elif [ "$service" = "gateway" ]; then
        log_info "更新网关服务..."
        cd ${DEPLOY_PATH}
        export BLOG_PASSWORD="$BLOG_PASSWORD"
        ./deploy/deploy-gateway-service.sh
    elif [ "$service" = "common" ]; then
        log_info "更新公共模块..."
        cd ${DEPLOY_PATH}/common
        silent_exec go mod download
        silent_exec go mod tidy
        log_info "公共模块更新完成"
    else
        log_error "未知服务: $service"
        exit 1
    fi
}

# 主函数
main() {
    # 检查是否为root用户
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用root权限运行此脚本"
        exit 1
    fi
    
    # 检查参数
    if [ $# -eq 0 ]; then
        show_help
        exit 0
    fi
    
    # 解析参数
    while [ $# -gt 0 ]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -l|--list)
                list_services
                exit 0
                ;;
            -s|--status)
                show_status "$2"
                exit 0
                ;;
            -r|--restart)
                if [ -z "$2" ]; then
                    log_error "请指定要重启的服务"
                    exit 1
                fi
                restart_service "$2"
                exit 0
                ;;
            -t|--stop)
                if [ -z "$2" ]; then
                    log_error "请指定要停止的服务"
                    exit 1
                fi
                stop_service "$2"
                exit 0
                ;;
            -u|--update)
                if [ -z "$2" ]; then
                    log_error "请指定要更新的服务"
                    exit 1
                fi
                update_service "$2"
                exit 0
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
        shift
    done
}

# 执行主函数
main "$@" 