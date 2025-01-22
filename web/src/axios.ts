import a, {
  AxiosError,
  AxiosRequestConfig,
  InternalAxiosRequestConfig,
} from 'axios';
import { refreshAccessToken } from './api/auth';
import { API_BASE_URL } from './constants';

export const axios = a.create({
  baseURL: API_BASE_URL,
});

axios.interceptors.request.use(function (config: InternalAxiosRequestConfig) {
  const token = window.sessionStorage.getItem('access-token');
  if (token) {
    config.headers.Authorization = `Bearer ${JSON.parse(token).access_token}`;
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
  function (response) {
    return response;
  },
  async function (error) {
    const originalRequest: AxiosRequestConfig = error.config;
    if (!isRefreshing) {
      isRefreshing = true;
      try {
        const newAccessToken = await refreshAccessToken();

        error.config.headers['Authorization'] = `Bearer ${newAccessToken}`;

        refreshAndRetryQueue.forEach(({ config, resolve, reject }) => {
          axios
            .request(config)
            .then((response) => resolve(response))
            .catch((err) => reject(err));
        });

        refreshAndRetryQueue.length = 0;

        return axios(originalRequest);
      } catch (error) {
        throw new Error((error as AxiosError).message);
      } finally {
        isRefreshing = false;
      }
    }

    return new Promise<unknown>((resolve, reject) => {
      refreshAndRetryQueue.push({ config: originalRequest, resolve, reject });
    });
  }
);
