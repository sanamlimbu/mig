import a, {
  AxiosError,
  AxiosRequestConfig,
  InternalAxiosRequestConfig,
} from 'axios';
import { refreshAccessToken } from './api/auth';
import { API_BASE_URL } from './constants';
import { getAuthToken, removeAuthToken, setAuthToken } from './utils/auth';

export const axios = a.create({
  baseURL: API_BASE_URL,
  withCredentials: true,
});

axios.interceptors.request.use(function (config: InternalAxiosRequestConfig) {
  const authToken = getAuthToken();
  if (authToken) {
    config.headers.Authorization = `Bearer ${authToken.access_token}`;
  }

  config.headers['Content-Type'] = 'application/json';

  return config;
});

interface RetryQueueItem {
  resolve: (value?: unknown) => void;
  reject: (error?: unknown) => void;
  config: AxiosRequestConfig;
}

const refreshAndRetryQueue: RetryQueueItem[] = [];

let isRefreshing = false;

axios.interceptors.response.use(
  (response) => response,
  async function (error) {
    const originalRequest: AxiosRequestConfig = error.config;

    if (error.response?.status !== 401) {
      return Promise.reject(error);
    }

    if (originalRequest.url?.includes('/auth/')) {
      removeAuthToken();
      window.location.reload();
      return Promise.reject(error);
    }

    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        refreshAndRetryQueue.push({ config: originalRequest, resolve, reject });
      });
    }

    isRefreshing = true;

    try {
      const authToken = await refreshAccessToken();
      setAuthToken(authToken);

      refreshAndRetryQueue.forEach(({ config, resolve, reject }) => {
        if (config.headers) {
          config.headers['Authorization'] = `Bearer ${authToken.access_token}`;
        }

        axios
          .request(config)
          .then((response) => resolve(response))
          .catch((err) => reject(err));
      });

      refreshAndRetryQueue.length = 0;

      if (originalRequest.headers) {
        originalRequest.headers[
          'Authorization'
        ] = `Bearer ${authToken.access_token}`;
      }

      return axios(originalRequest);
    } catch (error) {
      throw new Error((error as AxiosError).message);
    } finally {
      isRefreshing = false;
    }
  }
);
