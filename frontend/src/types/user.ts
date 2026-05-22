export interface User {
  id: number;
  email: string;
  display_name: string;
  role: "admin" | "user";
  status: "active" | "disabled";
  last_login_at?: string;
  created_at: string;
  updated_at: string;
}

export interface UserInput {
  email: string;
  password?: string;
  display_name: string;
  role: User["role"];
  status: User["status"];
}
