import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { sendOtpCode, verifyOtpCode } from '/src/api/gameApi.jsx';
import './main.css';

// ... (Функция validateName остается без изменений) ...
const validateName = (name) => {
  const errors = [];
  const trimmed = name.trim();
  if (!trimmed) { errors.push('ФИО обязательно для заполнения'); return errors; }
  const validPattern = /^[а-яА-ЯёЁa-zA-Z\s\-]+$/;
  if (!validPattern.test(trimmed)) { errors.push('ФИО должно содержать только буквы, пробелы и дефисы'); }
  const words = trimmed.split(/\s+/).filter(word => word.length > 0);
  if (words.length < 2) { errors.push('Введите имя и фамилию'); }
  if (words.some(word => word.length < 2)) { errors.push('Каждая часть ФИО должна содержать минимум 2 символа'); }
  if (trimmed.length > 100) { errors.push('ФИО не должно превышать 100 символов'); }
  if (/\d/.test(trimmed)) { errors.push('ФИО не должно содержать цифры'); }
  return errors;
};

function Registration() {
  const navigate = useNavigate();

  // Состояния полей
  const [loginData, setLoginData] = useState({ name: '', email: '', code: '' });
  
  // Состояния интерфейса
  const [isCodeSent, setIsCodeSent] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [serverError, setServerError] = useState('');
  const [resendCooldown, setResendCooldown] = useState(0);
  const cooldownRef = useRef(null);
  
  // Состояния ошибок валидации
  const [errors, setErrors] = useState({ name: [], email: [], code: [] });
  const [touched, setTouched] = useState({ name: false, email: false, code: false });


  const handleLoginChange = (e) => {
    const { name, value } = e.target;
    setLoginData({ ...loginData, [name]: value });
    if (name === 'email') { setIsCodeSent(false); setServerError(''); }
    if (touched[name] && name === 'name') {
      setErrors(prev => ({ ...prev, name: validateName(value) }));
    }
  };

  const handleBlur = (e) => {
    const { name, value } = e.target;
    setTouched(prev => ({ ...prev, [name]: true }));
    if (name === 'name') {
      setErrors(prev => ({ ...prev, name: validateName(value) }));
    }
  };

  const isFormValid = () => {
    const nameErrors = validateName(loginData.name);
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    let isValid = true;
    let newErrors = { ...errors, name: nameErrors };
    if (!emailRegex.test(loginData.email)) {
      newErrors.email = ['Некорректный формат email'];
      isValid = false;
    } else { newErrors.email = []; }
    if (nameErrors.length > 0) isValid = false;
    setErrors(newErrors);
    return isValid;
  };

  const startCooldown = () => {
    setResendCooldown(60);
    cooldownRef.current = setInterval(() => {
      setResendCooldown(prev => {
        if (prev <= 1) {
          clearInterval(cooldownRef.current);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
  };

  useEffect(() => () => clearInterval(cooldownRef.current), []);

  const handleSendCode = async () => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!loginData.email || !emailRegex.test(loginData.email)) {
      setErrors(prev => ({ ...prev, email: ['Введите корректный email'] }));
      setTouched(prev => ({ ...prev, email: true }));
      return;
    }
    setIsLoading(true);
    setServerError('');
    try {
      await sendOtpCode(loginData.email);
      setIsCodeSent(true);
      setLoginData(prev => ({ ...prev, code: '' }));
      startCooldown();
    } catch (err) {
      setServerError(err.message || 'Не удалось отправить код.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleLoginSubmit = async (e) => {
    e.preventDefault();
    if (!isFormValid()) { setTouched({ name: true, email: true, code: true }); return; }
    if (!isCodeSent) { setServerError('Сначала получите код подтверждения'); return; }
    if (!loginData.code.trim()) { setErrors(prev => ({ ...prev, code: ['Введите код'] })); return; }

    setIsLoading(true);
    setServerError('');
    try {
      await verifyOtpCode(loginData.name, loginData.email, loginData.code);
      navigate('/profile'); 
    } catch (err) {
      setServerError(err.message || 'Неверный код или ошибка сервера.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleGithubAuth = () => {
    window.location.href = '/api/v1/auth/login';
  };

  return (
    <div className="overlay1">
      <div className="card1">
        <h1>Вход / Регистрация</h1>

        {serverError && (
          <div style={{ color: '#e74c3c', marginBottom: '15px', fontSize: '14px', textAlign: 'center' }}>
            {serverError}
          </div>
        )}

        <form onSubmit={handleLoginSubmit} className="auth-form" noValidate>
          {/* Поле ФИО */}
          <div className="form-group">
            <input type="text" name="name" placeholder="Введите ФИО" value={loginData.name} onChange={handleLoginChange} onBlur={handleBlur} required disabled={isLoading} className={errors.name.length > 0 && touched.name ? 'input-error' : ''} />
            {errors.name.length > 0 && touched.name && (<ul className="error-list">{errors.name.map((err, i) => (<li key={i} className="error-item">• {err}</li>))}</ul>)}
          </div>
          
          {/* Поле Email */}
          <div className="form-group">
            <input type="email" name="email" placeholder="Введите почту" value={loginData.email} onChange={handleLoginChange} onBlur={() => setTouched(prev => ({ ...prev, email: true }))} required disabled={isLoading} className={errors.email.length > 0 && touched.email ? 'input-error' : ''} />
            {errors.email.length > 0 && touched.email && (<ul className="error-list">{errors.email.map((e,i)=><li key={i} className="error-item">• {e}</li>)}</ul>)}
          </div>
          
          {/* Поле Кода */}
          <div className="password-field-group">
            <div className="password-input-wrapper">
              <input type="text" name="code" placeholder="Введите код из письма" value={loginData.code} onChange={handleLoginChange} onBlur={() => setTouched(prev => ({ ...prev, code: true }))} required disabled={isLoading || !isCodeSent} className={`password-input ${errors.code.length > 0 && touched.code ? 'input-error' : ''}`} />
              <button type="button" className="btn-send-code-inline" onClick={handleSendCode} disabled={isLoading || (!isCodeSent ? false : resendCooldown > 0)} title={isCodeSent ? 'Отправить код повторно' : 'Отправить код на почту'}>
                {isLoading ? 'Отправка...' : (isCodeSent ? (resendCooldown > 0 ? `Повторно через ${resendCooldown}с` : 'Отправить повторно') : 'Отправить код')}
              </button>
            </div>
          </div>

          {isCodeSent && !serverError && (<span className="password-sent-hint">✓ Код отправлен! Проверьте почту.</span>)}

          {/* Кнопки действий */}
          <button className="btn-primary" type="submit" disabled={isLoading}>
            {isLoading ? 'Проверка...' : 'Войти'}
          </button>

          <button type="button" className="btn-secondary" onClick={handleGithubAuth} disabled={isLoading}>
            Через GitHub
          </button>
        </form>

        <p className="agreement1">
          Продолжая, вы соглашаетесь с <a href="#">Пользовательским соглашением</a>
        </p>
      </div>
    </div>
  );
}

export default Registration;