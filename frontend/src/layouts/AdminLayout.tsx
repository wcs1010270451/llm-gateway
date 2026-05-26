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
import { Button, ConfigProvider, Layout, Menu, Space, Typography } from "antd";
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
  const currentSection = navItems.find((item) => item.key === selectedKey)?.label ?? "概览";

  function logout() {
    clearSession();
    navigate("/login", { replace: true });
  }

  return (
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: "#237bb2",
          colorInfo: "#237bb2",
          colorLink: "#237bb2",
          colorText: "#223842",
          colorTextSecondary: "#60757d",
          colorBorderSecondary: "#d9e6ec",
          colorBgContainer: "#fcfeff",
          borderRadius: 10,
          controlHeight: 40,
        },
        components: {
          Button: { borderRadius: 9, primaryShadow: "none" },
          Layout: { bodyBg: "#f4f9fc", headerBg: "#fcfeff", siderBg: "#fcfeff" },
          Menu: { itemBorderRadius: 11, itemHeight: 46 },
          Table: { headerBg: "#fcfeff", headerColor: "#60757d", rowHoverBg: "#eef7fb" },
        },
      }}
    >
      <Layout className="admin-shell">
        <a className="skip-link" href="#admin-main">
          跳到主要内容
        </a>
        <Layout.Sider width={248} collapsedWidth={76} collapsed={collapsed} className="admin-sider">
          <div className="admin-brand">
            <div className="admin-brand-mark">LG</div>
            {!collapsed ? (
              <div>
                <Typography.Text className="admin-brand-title">LLM Gateway</Typography.Text>
                <Typography.Text className="admin-brand-subtitle">Operations workspace</Typography.Text>
              </div>
            ) : null}
          </div>
          {!collapsed ? <Typography.Text className="admin-nav-label">管理功能</Typography.Text> : null}
          <Menu
            theme="light"
            className="admin-nav"
            mode="inline"
            selectedKeys={[selectedKey]}
            items={navItems}
            onClick={(item) => navigate(item.key)}
          />
        </Layout.Sider>
        <Layout className="admin-workspace">
          <Layout.Header className="admin-header">
            <Space size={16}>
              <Button
                className="admin-collapse"
                type="text"
                icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                onClick={() => setCollapsed(!collapsed)}
                aria-label={collapsed ? "展开导航" : "收起导航"}
              />
              <div className="admin-header-context">
                <Typography.Text className="admin-header-label">管理控制台</Typography.Text>
                <Typography.Text className="admin-header-title">{currentSection}</Typography.Text>
              </div>
            </Space>
            <Space className="admin-account" size={12}>
              <Typography.Text className="admin-email">{user?.email}</Typography.Text>
              <Button type="text" icon={<LogoutOutlined />} onClick={logout} aria-label="退出登录">
                退出
              </Button>
            </Space>
          </Layout.Header>
          <Layout.Content className="admin-content">
            <main id="admin-main" className="admin-main" tabIndex={-1}>
              <Outlet />
            </main>
          </Layout.Content>
        </Layout>
      </Layout>
    </ConfigProvider>
  );
}
