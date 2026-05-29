import { KeyOutlined } from "@ant-design/icons";
import { Button, ConfigProvider, Layout } from "antd";
import { Outlet, useLocation, useNavigate } from "react-router";

import { AccountNav } from "../components/AccountNav";
import { AppBrand } from "../components/AppBrand";

export function PortalLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const keysActive = location.pathname.startsWith("/portal/keys");

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
            <AppBrand variant="portal" subtitle="Access workspace" />
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
            <AccountNav variant="portal" />
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
