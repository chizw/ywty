import axios from 'axios';

const http = axios.create({ baseURL: '/api/admin', timeout: 15000, withCredentials: true });

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
