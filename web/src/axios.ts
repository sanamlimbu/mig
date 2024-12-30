import a, { InternalAxiosRequestConfig } from 'axios';
import { API_BASE_URL } from './constants';

export const axios = a.create({
  baseURL: API_BASE_URL,
});

axios.interceptors.request.use(function (config: InternalAxiosRequestConfig) {
  const token = localStorage.getItem('mig-auth-token');
  if (token) {
    config.headers.Authorization = `Bearer ${JSON.parse(token).access_token}`;
  }

  config.headers['Content-Type'] = 'application/json';

  return config;
});
