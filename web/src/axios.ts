import a from 'axios';
import { API_BASE_URL } from './constants';

export const axios = a.create({
  baseURL: API_BASE_URL,
});

