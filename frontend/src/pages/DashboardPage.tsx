import { useQuery } from "@tanstack/react-query";
import { Card, Skeleton, Table, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";

import { fetchAdminStats } from "../api/admin";
import { PageHeader } from "../components/PageHeader";
import type { KeyModelUsageStat } from "../types";

function formatNumber(value?: number) {
  return new Intl.NumberFormat("zh-CN").format(value ?? 0);
}

function formatPercent(successCount?: number, requestCount?: number) {
  if (!requestCount) {
    return "-";
  }
  return `${((successCount ?? 0) / requestCount * 100).toFixed(1)}%`;
}

function formatCost(value?: number) {
  return new Intl.NumberFormat("zh-CN", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 4,
    maximumFractionDigits: 4,
  }).format(value ?? 0);
}

const topModelColumns: ColumnsType<KeyModelUsageStat> = [
  { title: "模型", dataIndex: "public_model_name", render: (value) => value || "-" },
  { title: "请求", dataIndex: "request_count", width: 100, render: formatNumber },
  { title: "Token", dataIndex: "total_tokens", width: 120, render: formatNumber },
  { title: "预估成本", dataIndex: "estimated_cost", width: 130, render: formatCost },
];

export function DashboardPage() {
  const statsQuery = useQuery({ queryKey: ["admin", "stats"], queryFn: fetchAdminStats });
  const stats = statsQuery.data;
  const cards = [
    { label: "累计请求", value: formatNumber(stats?.request_count), hint: "已进入网关" },
    { label: "累计 Token", value: formatNumber(stats?.total_tokens), hint: "输入与输出合计" },
    { label: "供应商通道", value: formatNumber(stats?.provider_count), hint: "已配置来源" },
    { label: "对外模型", value: formatNumber(stats?.model_count), hint: "可路由能力" },
  ];
  const recentCards = [
    { label: "请求", value: formatNumber(stats?.recent_usage?.request_count) },
    { label: "成功率", value: formatPercent(stats?.recent_usage?.success_count, stats?.recent_usage?.request_count) },
    { label: "活跃用户", value: formatNumber(stats?.recent_usage?.active_user_count) },
    { label: "活跃 Key", value: formatNumber(stats?.recent_usage?.active_key_count) },
    { label: "平均延迟", value: `${Math.round(stats?.recent_usage?.average_latency_ms ?? 0)} ms` },
    { label: "Token", value: formatNumber(stats?.recent_usage?.total_tokens) },
    { label: "预估成本", value: formatCost(stats?.recent_usage?.estimated_cost) },
  ];

  return (
    <div className="page-stack">
      <PageHeader eyebrow="OPERATIONS OVERVIEW" title="概览" description="查看网关运行状态、模型路由和近期调用情况。" />
      <Card className="admin-summary">
        {cards.map((card) => (
          <div className="admin-stat" key={card.label}>
            <Typography.Text className="admin-stat-label">{card.label}</Typography.Text>
            {statsQuery.isLoading ? (
              <Skeleton.Input active size="small" className="admin-stat-loading" />
            ) : (
              <Typography.Title level={2}>{card.value}</Typography.Title>
            )}
            <Typography.Text className="admin-stat-hint">{card.hint}</Typography.Text>
          </div>
        ))}
      </Card>
      <div className="admin-dashboard-grid">
        <Card
          className="admin-panel admin-recent-panel"
          title="近 24 小时运行态势"
          extra={<Typography.Text className="admin-panel-note">滚动统计窗口</Typography.Text>}
        >
          <div className="admin-recent-summary">
            {recentCards.map((card) => (
              <div className="admin-recent-stat" key={card.label}>
                <Typography.Text className="admin-recent-label">{card.label}</Typography.Text>
                {statsQuery.isLoading ? (
                  <Skeleton.Input active size="small" />
                ) : (
                  <Typography.Text className="admin-recent-value">{card.value}</Typography.Text>
                )}
              </div>
            ))}
          </div>
        </Card>
        <Card className="admin-panel admin-table-panel" title="近 24 小时热门模型">
          <Table
            className="admin-table"
            rowKey="public_model_name"
            columns={topModelColumns}
            dataSource={stats?.top_models ?? []}
            loading={statsQuery.isLoading}
            pagination={false}
            size="middle"
            scroll={{ x: "max-content" }}
          />
        </Card>
      </div>
    </div>
  );
}
