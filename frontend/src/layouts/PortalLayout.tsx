import { KeyOutlined, LogoutOutlined } from "@ant-design/icons";
import { Button, Layout, Menu, Space, Typography } from "antd";
import { Outlet, useLocation, useNavigate } from "react-router";

import { useAuthStore } from "../store/authStore";

const navItems = [{ key: "/portal/keys", icon: <KeyOutlined />, label: "我的 Key" }];

export function PortalLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, clearSession } = useAuthStore();

  function logout() {
    clearSession();
    navigate("/login", { replace: true });
  }

  return (
    <Layout className="app-shell">
      <Layout.Sider width={232} className="app-sider">
        <div className="brand">
          <div className="brand-mark">LG</div>
          <div>
            <Typography.Text className="brand-title">LLM Gateway</Typography.Text>
            <Typography.Text className="brand-subtitle">User Console</Typography.Text>
          </div>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[navItems.find((item) => location.pathname.startsWith(item.key))?.key ?? "/portal/keys"]}
          items={navItems}
          onClick={(item) => navigate(item.key)}
        />
      </Layout.Sider>
      <Layout>
        <Layout.Header className="app-header">
          <Typography.Text strong>用户控制台</Typography.Text>
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
