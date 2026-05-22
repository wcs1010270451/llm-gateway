import { Button, Card, Form, Input, Typography, message } from "antd";
import { useMutation } from "@tanstack/react-query";
import { Navigate, useLocation, useNavigate } from "react-router";

import { login } from "../api/auth";
import { useAuthStore } from "../store/authStore";

interface LoginFormValues {
  email: string;
  password: string;
}

export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { token, user, setSession } = useAuthStore();

  const mutation = useMutation({
    mutationFn: (values: LoginFormValues) => login(values.email, values.password),
    onSuccess: (result) => {
      setSession(result.token, result.user);
      const fallback = result.user.role === "admin" ? "/dashboard" : "/portal/keys";
      navigate((location.state as { from?: string } | null)?.from ?? fallback, { replace: true });
    },
    onError: (error) => message.error(error.message),
  });

  if (token && user) {
    return <Navigate to={user.role === "admin" ? "/dashboard" : "/portal/keys"} replace />;
  }

  return (
    <div className="login-page">
      <Card className="login-card">
        <div className="login-header">
          <Typography.Title level={2}>LLM Gateway</Typography.Title>
          <Typography.Text type="secondary">登录后进入对应控制台</Typography.Text>
        </div>
        <Form layout="vertical" onFinish={(values) => mutation.mutate(values as LoginFormValues)}>
          <Form.Item name="email" label="邮箱" rules={[{ required: true, message: "请输入邮箱" }]}>
            <Input autoComplete="email" />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true, message: "请输入密码" }]}>
            <Input.Password autoComplete="current-password" />
          </Form.Item>
          <Button type="primary" htmlType="submit" block loading={mutation.isPending}>
            登录
          </Button>
        </Form>
      </Card>
    </div>
  );
}
