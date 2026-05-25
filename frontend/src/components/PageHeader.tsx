import { Typography } from "antd";
import type { ReactNode } from "react";

interface PageHeaderProps {
  eyebrow?: string;
  title: string;
  description?: string;
  actions?: ReactNode;
}

export function PageHeader({ eyebrow, title, description, actions }: PageHeaderProps) {
  return (
    <div className="page-header">
      <div className="page-header-copy">
        {eyebrow ? <Typography.Text className="page-eyebrow">{eyebrow}</Typography.Text> : null}
        <Typography.Title level={1}>{title}</Typography.Title>
        {description ? <Typography.Paragraph className="page-description">{description}</Typography.Paragraph> : null}
      </div>
      {actions ? <div className="page-header-actions">{actions}</div> : null}
    </div>
  );
}
