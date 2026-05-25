import { KeyOutlined, LogoutOutlined } from "@ant-design/icons";
import { Button, ConfigProvider, Layout, Space, Typography } from "antd";
import { Outlet, useLocation, useNavigate } from "react-router";

import { useAuthStore } from "../store/authStore";

export function PortalLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, clearSession } = useAuthStore();
  const keysActive = location.pathname.startsWith("/portal/keys");

  function logout() {
    clearSession();
    navigate("/login", { replace: true });
  }

  return (
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: "#237bb2",
          colorText: "#223840",
          colorTextSecondary: "#60757d",
          colorBorderSecondary: "#d9e6ec",
          colorBgContainer: "#fcfeff",
          borderRadius: 8,
          borderRadiusLG: 18,
          controlHeight: 42,
        },
      }}
    >
      <Layout className="portal-shell">
        <a className="skip-link" href="#portal-main">
          跳至主要内容
        </a>
        <header className="portal-header">
          <div className="portal-header-inner">
            <button type="button" className="portal-brand" onClick={() => navigate("/portal/keys")} aria-label="LLM Gateway 门户首页">
              <span className="portal-brand-mark" aria-hidden="true">
                LG
              </span>
              <span>
                <Typography.Text className="portal-brand-title">LLM Gateway</Typography.Text>
                <Typography.Text className="portal-brand-subtitle">Access workspace</Typography.Text>
              </span>
            </button>
            <nav className="portal-nav" aria-label="门户导航">
              <Button
                type="text"
                icon={<KeyOutlined />}
                className={keysActive ? "portal-nav-item is-active" : "portal-nav-item"}
                onClick={() => navigate("/portal/keys")}
              >
                访问凭据
              </Button>
            </nav>
            <Space className="portal-account" size={12}>
              <Typography.Text className="portal-email">{user?.email}</Typography.Text>
              <Button className="portal-logout" type="text" icon={<LogoutOutlined />} onClick={logout} aria-label="退出登录">
                退出
              </Button>
            </Space>
          </div>
        </header>
        <Layout.Content className="portal-content">
          <main id="portal-main" className="portal-main" tabIndex={-1}>
            <Outlet />
          </main>
        </Layout.Content>
      </Layout>
    </ConfigProvider>
  );
}
