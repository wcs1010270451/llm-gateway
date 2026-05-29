import { Typography } from "antd";
import { useNavigate } from "react-router";

type BrandVariant = "landing" | "portal" | "admin";

interface AppBrandProps {
  variant: BrandVariant;
  subtitle: string;
  showText?: boolean;
}

export function AppBrand({ variant, subtitle, showText = true }: AppBrandProps) {
  const navigate = useNavigate();

  return (
    <button type="button" className={`${variant}-brand`} onClick={() => navigate("/")} aria-label="LLM Gateway 首页">
      <span className={`${variant}-brand-mark`} aria-hidden="true">
        LG
      </span>
      {showText ? (
        <span>
          <Typography.Text className={`${variant}-brand-title`}>LLM Gateway</Typography.Text>
          <Typography.Text className={`${variant}-brand-subtitle`}>{subtitle}</Typography.Text>
        </span>
      ) : null}
    </button>
  );
}
