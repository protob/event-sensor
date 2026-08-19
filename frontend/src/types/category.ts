export interface Category {
  id: string;
  name: string;
  artist_count: number;
  created_at: string;
  updated_at: string;
}

export interface CreateCategoryPayload {
  name: string;
  user_id: string;
}

export interface UpdateCategoryPayload {
  name: string;
}
