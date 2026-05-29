import { ApiOutlined, ArrowRightOutlined, BarChartOutlined, KeyOutlined, SafetyCertificateOutlined } from "@ant-design/icons";
import { Button, Typography } from "antd";
import { useNavigate } from "react-router";

import { AccountNav } from "../components/AccountNav";
import { AppBrand } from "../components/AppBrand";
import { useAuthStore } from "../store/authStore";

const capabilities = [
  {
    title: "统一接入",
    text: "把 OpenAI、Anthropic、Gemini、Vertex AI 和自定义上游收束成一个入口。",
  },
  {
    title: "手动路由",
    text: "模型与供应商映射由管理员明确选择，变更路径清楚可追踪。",
  },
  {
    title: "凭据工作台",
    text: "用户登录后创建 Key、复制密钥、查看额度、追踪每次调用。",
  },
];

const signals = [
  { label: "Gateway", value: "统一协议" },
  { label: "Keys", value: "加密保存" },
  { label: "Logs", value: "请求可追踪" },
];

export function LandingPage() {
  const navigate = useNavigate();
  const { token, user } = useAuthStore();
  const isSignedIn = Boolean(token && user);
  const primaryEntry = user?.role === "admin" ? { label: "进入管理端", path: "/dashboard" } : { label: "查看凭据门户", path: "/portal/keys" };

  return (
    <main className="landing-shell">
      <header className="landing-nav" aria-label="官网导航">
        <AppBrand variant="landing" subtitle="Model access ledger" />
        <nav className="landing-nav-actions" aria-label="入口">
          <AccountNav variant="landing" showLogin />
        </nav>
      </header>

      <section className="landing-hero">
        <div className="landing-hero-copy">
          <Typography.Text className="landing-kicker">A quiet control plane for model access</Typography.Text>
          <Typography.Title>把多模型上游整理成一套可控的访问系统。</Typography.Title>
          <Typography.Paragraph>
            LLM Gateway 为团队提供统一调用入口、明确的模型路由、加密凭据与可追踪日志。它不是聊天产品，而是一张用于管理模型访问的清晰账本。
          </Typography.Paragraph>
          <div className="landing-hero-actions">
            <Button type="primary" size="large" icon={<ArrowRightOutlined />} onClick={() => navigate(isSignedIn ? primaryEntry.path : "/login")}>
              {isSignedIn ? primaryEntry.label : "登录进入"}
            </Button>
          </div>
        </div>

        <div className="landing-map" aria-label="LLM Gateway 能力示意">
          <div className="landing-map-core">
            <span>LLM Gateway</span>
            <small>routing · keys · logs</small>
          </div>
          <div className="landing-map-ring landing-map-ring-one" />
          <div className="landing-map-ring landing-map-ring-two" />
          <div className="landing-map-node node-a">OpenAI</div>
          <div className="landing-map-node node-b">Claude</div>
          <div className="landing-map-node node-c">Gemini</div>
          <div className="landing-map-node node-d">Vertex AI</div>
        </div>
      </section>

      <section className="landing-signals" aria-label="核心信号">
        {signals.map((item) => (
          <div className="landing-signal" key={item.label}>
            <span>{item.label}</span>
            <strong>{item.value}</strong>
          </div>
        ))}
      </section>

      <section className="landing-section">
        <div className="landing-section-copy">
          <Typography.Text className="landing-kicker">What it keeps clear</Typography.Text>
          <Typography.Title level={2}>访问、路由、日志，各自有边界。</Typography.Title>
        </div>
        <div className="landing-capabilities">
          {capabilities.map((item, index) => (
            <article className="landing-capability" key={item.title}>
              <span className="landing-capability-index">{String(index + 1).padStart(2, "0")}</span>
              <Typography.Title level={3}>{item.title}</Typography.Title>
              <Typography.Paragraph>{item.text}</Typography.Paragraph>
            </article>
          ))}
        </div>
      </section>

      <section className="landing-flow" aria-label="管理流程">
        <div className="landing-flow-step">
          <SafetyCertificateOutlined />
          <span>上游密钥加密保存</span>
        </div>
        <div className="landing-flow-line" />
        <div className="landing-flow-step">
          <ApiOutlined />
          <span>模型路由明确选择</span>
        </div>
        <div className="landing-flow-line" />
        <div className="landing-flow-step">
          <KeyOutlined />
          <span>用户 Key 独立管理</span>
        </div>
        <div className="landing-flow-line" />
        <div className="landing-flow-step">
          <BarChartOutlined />
          <span>日志与用量可追踪</span>
        </div>
      </section>

      <footer className="landing-footer">
        <Typography.Text>LLM Gateway</Typography.Text>
        <Button type="link" onClick={() => navigate(isSignedIn ? primaryEntry.path : "/login")}>
          {isSignedIn ? primaryEntry.label : "登录进入"}
        </Button>
      </footer>
    </main>
  );
}
