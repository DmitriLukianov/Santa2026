export const BASE_URL = '/api/v1';

export const getHeaders = () => {
  const token = localStorage.getItem('token');
  return {
    'Content-Type': 'application/json',
    ...(token && { Authorization: `Bearer ${token}` }),
  };
};

export const handleResponse = async (response) => {
  if (!response.ok) {
    // 401 — токен невалиден или пользователь не найден в БД
    if (response.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/registration';
      throw new Error('Сессия истекла. Войдите снова.');
    }
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.message || errorData.error || `Ошибка HTTP: ${response.status}`);
  }
  if (response.status === 204) return null;
  const text = await response.text();
  if (!text) return null;
  return JSON.parse(text);
};
