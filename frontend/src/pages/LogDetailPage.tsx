import { ArrowLeftOutlined } from "@ant-design/icons";
import { useQuery } from "@tanstack/react-query";
import { Button, Card, Typography } from "antd";
import { useNavigate, useParams } from "react-router";

import { fetchLog } from "../api/admin";
import { PageHeader } from "../components/PageHeader";
import { RequestLogDetailView } from "../components/RequestLogDetailView";

export function LogDetailPage() {
  const navigate = useNavigate();
  const params = useParams();
  const logID = Number(params.id);
  const logQuery = useQuery({
    queryKey: ["admin", "logs", logID],
    queryFn: () => fetchLog(logID),
    enabled: Number.isFinite(logID) && logID > 0,
  });

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="REQUEST TRACE"
        title="Log Detail"
        description="Full request and response payloads are loaded only on this page."
        actions={
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate("/logs")}>
            Back
          </Button>
        }
      />
      <Card className="admin-panel">
        {logQuery.isLoading || !logQuery.data ? (
          <Typography.Text type="secondary">Loading...</Typography.Text>
        ) : (
          <RequestLogDetailView item={logQuery.data} showUser />
        )}
      </Card>
    </div>
  );
}
