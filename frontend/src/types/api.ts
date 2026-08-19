export interface ApiError {
  title: string;
  status: number;
  detail?: string;
  errors?: ValidationError[];
}

export interface ValidationError {
  message: string;
  location: string;
  value: unknown;
}

export interface MessageResponse {
  message: string;
}

export interface SettingEntry {
  key: string;
  value: string;
  updated_at?: string;
}

export interface EventFilterParams {
  start_date?: string;
  end_date?: string;
  q?: string;
}
