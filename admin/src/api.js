import axios from 'axios';

const http = axios.create({ baseURL: '/api/admin', timeout: 20000, withCredentials: true });

http.interceptors.response.use(
  (r) => r,
  (e) => {
    const status = e.response?.status;
    if ((status === 401 || status === 403) && !location.pathname.endsWith('/login')) {
      localStorage.removeItem('admin_logged_in');
      window.location.href = '/admin/login';
    }
    return Promise.reject(e);
  },
);

export async function login(username, password) {
  const { data } = await http.post('/login', { username, password });
  return data;
}

export async function get(url) {
  const { data } = await http.get(url);
  return data;
}

export async function post(url, body) {
  const { data } = await http.post(url, body);
  return data;
}

export async function put(url, body) {
  const { data } = await http.put(url, body);
  return data;
}

export async function del(url) {
  const { data } = await http.delete(url);
  return data;
}

export function errMsg(e) {
  return e?.response?.data?.message || e?.message || '请求失败';
}
