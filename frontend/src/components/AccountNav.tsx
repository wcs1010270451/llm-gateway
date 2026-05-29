import { LogoutOutlined } from "@ant-design/icons";
import { Button, Space, Typography } from "antd";
import { useNavigate } from "react-router";

import { useAuthStore } from "../store/authStore";

type AccountVariant = "landing" | "portal" | "admin";

interface AccountNavProps {
  variant: AccountVariant;
  showLogin?: boolean;
}

export function AccountNav({ variant, showLogin = false }: AccountNavProps) {
  const navigate = useNavigate();
  const { token, user, clearSession } = useAuthStore();
  const isSignedIn = Boolean(token && user);

  function logout() {
    clearSession();
    navigate("/login", { replace: true });
  }

  if (!isSignedIn) {
    return showLogin ? (
      <Button type="primary" onClick={() => navigate("/login")}>
        登录
      </Button>
    ) : null;
  }

  return (
    <Space className={`${variant}-account`} size={12}>
      <Typography.Text className={`${variant}-email`}>{user?.email}</Typography.Text>
      <Button className={`${variant}-logout`} type="text" icon={<LogoutOutlined />} onClick={logout} aria-label="退出登录">
        退出
      </Button>
    </Space>
  );
}
