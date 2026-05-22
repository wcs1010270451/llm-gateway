import {
  ApiOutlined,
  BarChartOutlined,
  ClusterOutlined,
  DatabaseOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  PartitionOutlined,
  UserOutlined,
} from "@ant-design/icons";
import { Button, Layout, Menu, Space, Typography } from "antd";
import { Outlet, useLocation, useNavigate } from "react-router";

import { useAuthStore } from "../store/authStore";
import { useUIStore } from "../store/uiStore";

const navItems = [
  { key: "/dashboard", icon: <BarChartOutlined />, label: "概览" },
  { key: "/providers", icon: <DatabaseOutlined />, label: "供应商" },
  { key: "/models", icon: <ClusterOutlined />, label: "模型" },
  { key: "/model-families", icon: <PartitionOutlined />, label: "模型系列" },
  { key: "/users", icon: <UserOutlined />, label: "用户" },
  { key: "/logs", icon: <ApiOutlined />, label: "日志" },
];

export function AdminLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { collapsed, setCollapsed } = useUIStore();
  const { user, clearSession } = useAuthStore();

  const selectedKey = navItems.find((item) => location.pathname.startsWith(item.key))?.key ?? "/dashboard";

  function logout() {
    clearSession();
    navigate("/login", { replace: true });
  }

  return (
    <Layout className="app-shell">
      <Layout.Sider width={232} collapsedWidth={72} collapsed={collapsed} className="app-sider">
        <div className="brand">
          <div className="brand-mark">LG</div>
          {!collapsed ? (
            <div>
              <Typography.Text className="brand-title">LLM Gateway</Typography.Text>
              <Typography.Text className="brand-subtitle">Admin Console</Typography.Text>
            </div>
          ) : null}
        </div>
        <Menu theme="dark" mode="inline" selectedKeys={[selectedKey]} items={navItems} onClick={(item) => navigate(item.key)} />
      </Layout.Sider>
      <Layout>
        <Layout.Header className="app-header">
          <Space>
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
              aria-label={collapsed ? "展开导航" : "收起导航"}
            />
            <Typography.Text strong>模型上游控制台</Typography.Text>
          </Space>
          <Space className="header-user">
            <Typography.Text type="secondary">{user?.email}</Typography.Text>
            <Button type="text" icon={<LogoutOutlined />} onClick={logout} />
          </Space>
        </Layout.Header>
        <Layout.Content className="app-content">
          <Outlet />
        </Layout.Content>
      </Layout>
    </Layout>
  );
}
