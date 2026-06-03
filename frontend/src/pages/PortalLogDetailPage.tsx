import { ArrowLeftOutlined } from "@ant-design/icons";
import { useQuery } from "@tanstack/react-query";
import { Button, Card, Typography } from "antd";
import { useNavigate, useParams } from "react-router";

import { fetchMyLog } from "../api/apiKeys";
import { PageHeader } from "../components/PageHeader";
import { RequestLogDetailView } from "../components/RequestLogDetailView";

export function PortalLogDetailPage() {
  const navigate = useNavigate();
  const params = useParams();
  const keyID = Number(params.keyID);
  const logID = Number(params.logID);
  const logQuery = useQuery({
    queryKey: ["me", "logs", logID],
    queryFn: () => fetchMyLog(logID),
    enabled: Number.isFinite(logID) && logID > 0,
  });

  return (
    <div className="page-stack">
      <PageHeader
        eyebrow="REQUEST TRACE"
        title="Log Detail"
        description="Full request and response payloads are loaded only on this page."
        actions={
          <Button icon={<ArrowLeftOutlined />} onClick={() => navigate(`/portal/keys/${keyID}`)}>
            Back
          </Button>
        }
      />
      <Card className="portal-panel">
        {logQuery.isLoading || !logQuery.data ? (
          <Typography.Text type="secondary">Loading...</Typography.Text>
        ) : (
          <RequestLogDetailView item={logQuery.data} />
        )}
      </Card>
    </div>
  );
}
