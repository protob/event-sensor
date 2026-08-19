export interface User {
  id: string;
  username: string;
  email: string;
  role: string;
  active: boolean;
  created_at: string;
  updated_at: string;
}

export interface LoginPayload {
  username: string;
  password: string;
}

export interface AuthResponse {
  user: User;
  token: string;
}

export interface ChangePasswordPayload {
  current_password: string;
  new_password: string;
}

export interface UpdateProfilePayload {
  username?: string;
  email?: string;
}

export interface ResetPasswordPayload {
  username: string;
  email: string;
  new_password: string;
}
