import { useQuery } from "@tanstack/react-query";
import { Card, Table, Tag, Typography } from "antd";
import type { ColumnsType } from "antd/es/table";
import { useState } from "react";
import { useNavigate } from "react-router";

import { fetchLogs } from "../api/admin";
import { PageHeader } from "../components/PageHeader";
import type { RequestLog } from "../types";

function formatDateTime(value?: string) {
  return value ? new Date(value).toLocaleString("zh-CN") : "-";
}

function formatNumber(value?: number) {
  return new Intl.NumberFormat("zh-CN").format(value ?? 0);
}

export function LogsPage() {
  const navigate = useNavigate();
  const [pageState, setPageState] = useState({ page: 1, pageSize: 20 });
  const logsQuery = useQuery({
    queryKey: ["admin", "logs", pageState],
    queryFn: () => fetchLogs({ page: pageState.page, page_size: pageState.pageSize }),
  });

  const columns: ColumnsType<RequestLog> = [
    { title: "时间", dataIndex: "created_at", width: 180, render: formatDateTime },
    {
      title: "用户",
      key: "user",
      width: 200,
      render: (_, record) => record.user?.display_name || record.user?.email || "-",
    },
    { title: "模型", dataIndex: "public_model_name", width: 180, ellipsis: true, render: (value) => value || "-" },
    {
      title: "状态",
      dataIndex: "success",
      width: 90,
      render: (value) => <Tag color={value ? "success" : "error"}>{value ? "成功" : "失败"}</Tag>,
    },
    { title: "HTTP", dataIndex: "http_status", width: 90 },
    { title: "总 Token", dataIndex: "total_tokens", width: 120, render: formatNumber },
    { title: "延迟", dataIndex: "latency_ms", width: 110, render: (value) => `${value ?? 0} ms` },
    { title: "错误", dataIndex: "error_message", width: 240, ellipsis: true, render: (value) => value || "-" },
  ];

  return (
    <div className="page-stack">
      <PageHeader eyebrow="REQUEST TRACE" title="请求日志" description="列表只加载轻量字段，点击记录查看完整请求和响应。" />
      <Card className="admin-panel admin-table-panel" title="调用记录" extra={<Typography.Text className="admin-panel-note">选择一行打开详情页</Typography.Text>}>
        <Table
          className="admin-table"
          rowKey="id"
          columns={columns}
          dataSource={logsQuery.data?.items ?? []}
          loading={logsQuery.isLoading}
          scroll={{ x: 980 }}
          onRow={(record) => ({
            className: "admin-table-row-action",
            tabIndex: 0,
            onClick: () => navigate(`/logs/${record.id}`),
            onKeyDown: (event) => {
              if (event.key === "Enter" || event.key === " ") {
                event.preventDefault();
                navigate(`/logs/${record.id}`);
              }
            },
          })}
          pagination={{
            current: logsQuery.data?.page ?? pageState.page,
            pageSize: logsQuery.data?.page_size ?? pageState.pageSize,
            total: logsQuery.data?.total ?? 0,
            showSizeChanger: true,
          }}
          onChange={(pagination) =>
            setPageState({
              page: pagination.current ?? 1,
              pageSize: pagination.pageSize ?? 20,
            })
          }
        />
      </Card>
    </div>
  );
}
