export interface ModelFamily {
  id: number;
  name: string;
  display_name: string;
  status: "active" | "disabled";
  description: string;
  created_at: string;
  updated_at: string;
}

export interface ModelFamilyInput {
  name: string;
  display_name: string;
  status: ModelFamily["status"];
  description: string;
}
