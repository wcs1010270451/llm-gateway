import { useQuery } from "@tanstack/react-query";
import { Card, Col, Row, Skeleton, Typography } from "antd";

import { fetchAdminStats } from "../api/admin";
import { PageHeader } from "../components/PageHeader";

function formatNumber(value?: number) {
  return new Intl.NumberFormat("zh-CN").format(value ?? 0);
}

export function DashboardPage() {
  const statsQuery = useQuery({ queryKey: ["admin", "stats"], queryFn: fetchAdminStats });
  const stats = statsQuery.data;
  const cards = [
    { label: "请求数", value: formatNumber(stats?.request_count) },
    { label: "总 Token", value: formatNumber(stats?.total_tokens) },
    { label: "供应商", value: formatNumber(stats?.provider_count) },
    { label: "模型", value: formatNumber(stats?.model_count) },
  ];

  return (
    <div className="page-stack">
      <PageHeader title="概览" description="查看网关运行状态、模型路由和近期调用情况。" />
      <Row gutter={[16, 16]}>
        {cards.map((card) => (
          <Col xs={24} sm={12} lg={6} key={card.label}>
            <Card className="metric-card">
              <Typography.Text type="secondary">{card.label}</Typography.Text>
              {statsQuery.isLoading ? <Skeleton.Input active size="small" style={{ marginTop: 8 }} /> : <Typography.Title level={3}>{card.value}</Typography.Title>}
            </Card>
          </Col>
        ))}
      </Row>
    </div>
  );
}
