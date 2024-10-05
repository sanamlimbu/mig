import { createClient } from '@supabase/supabase-js';
import { useMemo } from 'react';
import { SUPABASE_ANON_KEY, SUPABASE_URL } from '../constants';

const supabase = createClient(SUPABASE_URL, SUPABASE_ANON_KEY);

export const useSupabase = () => {
  return useMemo(() => supabase, []);
};
