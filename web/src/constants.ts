export const WS_BASE_URL =
  import.meta.env.VITE_WS_BASE_URL || 'ws://127.0.0.1:8080/ws';

export const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:8080/api';

export const SUPABASE_URL = import.meta.env.VITE_SUPABASE_URL || '';
export const SUPABASE_ANON_KEY = import.meta.env.VITE_SUPABASE_ANON_KEY || '';
