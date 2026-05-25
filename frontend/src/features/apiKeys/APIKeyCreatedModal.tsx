import { CheckOutlined, CopyOutlined } from "@ant-design/icons";
import { Alert, Button, Input, Modal, Space, Typography, message } from "antd";
import { useState } from "react";

import type { CreatedAPIKey } from "../../types";

interface APIKeyCreatedModalProps {
  created?: CreatedAPIKey;
  onClose: () => void;
}

export function APIKeyCreatedModal({ created, onClose }: APIKeyCreatedModalProps) {
  const [copied, setCopied] = useState(false);

  async function copyKey() {
    if (!created) {
      return;
    }
    try {
      await navigator.clipboard.writeText(created.plain_key);
      setCopied(true);
      message.success("完整 Key 已复制");
    } catch {
      message.error("复制失败，请手动复制");
    }
  }

  function close() {
    setCopied(false);
    onClose();
  }

  return (
    <Modal
      className="portal-modal"
      open={Boolean(created)}
      title="Key 已创建"
      onCancel={close}
      footer={
        <Button type="primary" onClick={close}>
          关闭
        </Button>
      }
      destroyOnHidden
    >
      <Space direction="vertical" size={16} style={{ width: "100%" }}>
        <Alert
          type="info"
          showIcon
          message="请保存完整 Key。你之后也可以在凭据列表中点击复制，系统会按需安全读取。"
        />
        <div>
          <Typography.Text type="secondary">完整 Key</Typography.Text>
          <Input
            value={created?.plain_key ?? ""}
            readOnly
            style={{ marginTop: 8 }}
            suffix={
              <Button
                type="text"
                size="small"
                icon={copied ? <CheckOutlined /> : <CopyOutlined />}
                onClick={() => void copyKey()}
                aria-label="复制完整 Key"
              />
            }
          />
        </div>
      </Space>
    </Modal>
  );
}
