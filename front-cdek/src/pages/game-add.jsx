import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
// Импортируем метод создания игры
import { createGame, isAuthenticated } from '/src/api/gameApi.jsx';
import './main.css';

// === ФУНКЦИИ ВАЛИДАЦИИ (без изменений) ===

const validateTeamName = (name) => {
  const errors = [];
  const trimmed = name.trim();
  
  if (!trimmed) {
    errors.push('Название команды обязательно');
    return errors;
  }
  
  const validPattern = /^[а-яА-ЯёЁa-zA-Z0-9\s\-\,\.\(\)\/]+$/;
  if (!validPattern.test(trimmed)) {
    errors.push('Разрешены только буквы, цифры, пробел и символы: - , . ( ) /');
  }
  
  if (trimmed.length < 3) {
    errors.push('Минимальная длина названия — 3 символа');
  }
  if (trimmed.length > 150) {
    errors.push('Максимальная длина названия — 150 символов');
  }
  
  if (trimmed.startsWith(' ') || trimmed.endsWith(' ')) {
    errors.push('Название не должно начинаться или заканчиваться пробелом');
  }
  
  return errors;
};

const validateDrawDate = (dateString) => {
  const errors = [];

  if (!dateString) {
    errors.push('Дата жеребьёвки обязательна');
    return errors;
  }

  const date = new Date(dateString);
  if (isNaN(date.getTime())) {
    errors.push('Введите корректную дату');
    return errors;
  }

  const today = new Date();
  today.setHours(0, 0, 0, 0);
  date.setHours(0, 0, 0, 0);
  if (date < today) {
    errors.push('Дата жеребьёвки не может быть в прошлом');
  }

  return errors;
};

const validateBudget = (value) => {
  const errors = [];
  const num = Number(value);

  if (!value) {
    errors.push('Укажите бюджет подарка');
    return errors;
  }

  if (isNaN(num)) {
    errors.push('Бюджет должен быть числом');
    return errors;
  }

  if (num <= 0) {
    errors.push('Бюджет должен быть больше 0');
  }

  if (num > 5000) {
    errors.push('Бюджет слишком большой (макс. 5 000)');
  }

  return errors;
};

const validateOrganizerNotes = (notes) => {
  const errors = [];
  if (notes && notes.length > 500) {
    errors.push('Максимальная длина — 500 символов');
  }
  return errors;
};

function Game_add() {
  const navigate = useNavigate();

  React.useEffect(() => {
    if (!isAuthenticated()) {
      navigate('/registration', { replace: true });
    }
  }, [navigate]);
  
  // Состояния формы
  const [teamName, setTeamName] = useState('');
  const [drawDate, setDrawDate] = useState('');
  const [drawTime, setDrawTime] = useState('12:00');
  const [giftBudget, setGiftBudget] = useState('');
  const [wantParticipate, setWantParticipate] = useState(false);
  const [organizerNotes, setOrganizerNotes] = useState('');
  
  // Состояния ошибок и касаний
  const [errors, setErrors] = useState({ 
    teamName: [], 
    drawDate: [], 
    giftBudget: [], 
    organizerNotes: [] 
  });
  
  const [touched, setTouched] = useState({ 
    teamName: false, 
    drawDate: false, 
    giftBudget: false, 
    organizerNotes: false 
  });

  // ← НОВОЕ: Состояния загрузки и ошибки сервера
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState(null);

  const MIN_DATE = new Date().toISOString().split('T')[0];

  // --- Обработчики изменений (без изменений) ---

  const handleTeamNameChange = (e) => {
    const value = e.target.value;
    setTeamName(value);
    if (touched.teamName) {
      setErrors(prev => ({ ...prev, teamName: validateTeamName(value) }));
    }
  };

  const handleDateChange = (e) => {
    const value = e.target.value;
    setDrawDate(value);
    setErrors(prev => ({ ...prev, drawDate: validateDrawDate(value) }));
    setTouched(prev => ({ ...prev, drawDate: true }));
  };

  const handleBudgetChange = (e) => {
    const value = e.target.value;
    setGiftBudget(value);
    if (touched.giftBudget) {
      setErrors(prev => ({ ...prev, giftBudget: validateBudget(value) }));
    }
  };

  const handleNotesChange = (e) => {
    const value = e.target.value;
    setOrganizerNotes(value);
    if (touched.organizerNotes) {
      setErrors(prev => ({ ...prev, organizerNotes: validateOrganizerNotes(value) }));
    }
  };

  const handleBlur = (e) => {
    const { name, value } = e.target;
    setTouched(prev => ({ ...prev, [name]: true }));
    
    if (name === 'teamName') {
      setErrors(prev => ({ ...prev, teamName: validateTeamName(value) }));
    } else if (name === 'drawDate') {
      setErrors(prev => ({ ...prev, drawDate: validateDrawDate(value) }));
    } else if (name === 'giftBudget') {
      setErrors(prev => ({ ...prev, giftBudget: validateBudget(value) }));
    } else if (name === 'organizerNotes') {
      setErrors(prev => ({ ...prev, organizerNotes: validateOrganizerNotes(value) }));
    }
  };

  const isFormValid = () => {
    const nameErrors = validateTeamName(teamName);
    const dateErrors = validateDrawDate(drawDate);
    const budgetErrors = validateBudget(giftBudget);
    const notesErrors = validateOrganizerNotes(organizerNotes);
    
    setErrors({ 
      teamName: nameErrors, 
      drawDate: dateErrors, 
      giftBudget: budgetErrors, 
      organizerNotes: notesErrors 
    });
    
    return nameErrors.length === 0 && 
           dateErrors.length === 0 && 
           budgetErrors.length === 0 && 
           notesErrors.length === 0;
  };

  // ← НОВОЕ: Отправка данных на сервер
  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitError(null);
    
    if (!isFormValid()) {
      setTouched({ 
        teamName: true, 
        drawDate: true, 
        giftBudget: true, 
        organizerNotes: true 
      });
      return;
    }
    
    try {
      setIsSubmitting(true);

      // Подготовка данных для API
      const gameData = {
        title: teamName.trim(),
        drawDate: drawDate ? new Date(`${drawDate}T${drawTime}`).toISOString() : null,
        organizerNotes: organizerNotes.trim() || null,
        wantParticipate: wantParticipate,
      };

      // Вызов API создания игры
      const createdGame = await createGame(gameData);
      
      // Переход на страницу созданной игры
      // Предполагаем, что API возвращает объект с id: { id: '123', ... }
      const gameId = createdGame.id || createdGame.eventId;
      
      if (gameId) {
        navigate('/game/add/link', { state: { eventId: gameId } });
      } else {
        throw new Error('Не удалось получить ID созданной игры');
      }
      
    } catch (error) {
      console.error('Ошибка создания игры:', error);
      setSubmitError(error.message || 'Не удалось создать игру. Попробуйте позже.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleGoBack = () => {
    navigate(-1);
  };

  // Кнопка активна, если нет ошибок и форма не отправляется
  const canSubmit = teamName.trim() && 
                    drawDate && 
                    giftBudget && 
                    errors.teamName.length === 0 && 
                    errors.drawDate.length === 0 &&
                    errors.giftBudget.length === 0 &&
                    !isSubmitting;

  return (
    <div className="overlay_game_add"> 
      <div className="card_game_add"> 
        <h1>Создание игры</h1>

        <form onSubmit={handleSubmit} className="game-add-form" noValidate>                     
          
          {/* 1. Поле названия команды */}
          <div className="form-group">
            <label>Название команды <span className="required">*</span></label>
            <input 
              type="text"
              name="teamName"
              placeholder="Введите название команды"
              value={teamName}
              onChange={handleTeamNameChange}
              onBlur={handleBlur}
              disabled={isSubmitting}
              className={`input-field ${errors.teamName.length > 0 && touched.teamName ? 'input-error' : ''}`}
              required
            />
            {errors.teamName.length > 0 && touched.teamName && (
              <ul className="error-list">
                {errors.teamName.map((err, i) => (
                  <li key={i} className="error-item">• {err}</li>
                ))}
              </ul>
            )}
          </div>

          {/* 2. Поле выбора даты и времени */}
          <div className="form-group date-group">
            <label>Дата и время жеребьёвки <span className="required">*</span></label>
            <div style={{ display: 'flex', gap: '10px' }}>
              <input
                type="date"
                name="drawDate"
                value={drawDate}
                onChange={handleDateChange}
                onBlur={handleBlur}
                disabled={isSubmitting}
                min={MIN_DATE}
                className={`input-field date-input ${errors.drawDate.length > 0 && touched.drawDate ? 'input-error' : ''}`}
                style={{ flex: 2 }}
                required
              />
              <input
                type="time"
                value={drawTime}
                onChange={(e) => setDrawTime(e.target.value)}
                disabled={isSubmitting}
                className="input-field"
                style={{ flex: 1 }}
              />
            </div>
            {errors.drawDate.length > 0 && touched.drawDate && (
              <ul className="error-list">
                {errors.drawDate.map((err, i) => (
                  <li key={i} className="error-item">• {err}</li>
                ))}
              </ul>
            )}
          </div>

          {/* 3. ПОЛЕ ЦЕНЫ (БЮДЖЕТА) */}
          <div className="form-group">
            <label>Бюджет на подарок (руб.) <span className="required">*</span></label>
            <input 
              type="number"
              name="giftBudget"
              placeholder="Например: 1500"
              value={giftBudget}
              onChange={handleBudgetChange}
              onBlur={handleBlur}
              disabled={isSubmitting}
              min="1"
              step="100"
              className={`input-field ${errors.giftBudget.length > 0 && touched.giftBudget ? 'input-error' : ''}`}
              required
            />
            {errors.giftBudget.length > 0 && touched.giftBudget && (
              <ul className="error-list">
                {errors.giftBudget.map((err, i) => (
                  <li key={i} className="error-item">• {err}</li>
                ))}
              </ul>
            )}
          </div>

          {/* 4. Пожелания организатора */}
          <div className="form-group">
            <label>Пожелания от организатора <br /> <span style={{fontWeight: 'normal', fontSize: '12px', color: '#757575'}}>(отобразится в письмах участников после жеребьевки)</span></label>
            <textarea
              name="organizerNotes"
              placeholder="Например: Сбор подарков в офисе на 3 этаже 28.12..."
              value={organizerNotes}
              onChange={handleNotesChange}
              onBlur={handleBlur}
              disabled={isSubmitting}
              className={`input-field input-notes ${errors.organizerNotes.length > 0 && touched.organizerNotes ? 'input-error' : ''}`}
              rows={4}
              maxLength={500}
            />
            {errors.organizerNotes.length > 0 && touched.organizerNotes && (
              <ul className="error-list">
                {errors.organizerNotes.map((err, i) => (
                  <li key={i} className="error-item">• {err}</li>
                ))}
              </ul>
            )}
            <div className="notes-hint">{organizerNotes.length} / 500</div>
          </div>

          {/* Чекбокс участия */}
          <div className="form-group checkbox-group">
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={wantParticipate}
                onChange={(e) => setWantParticipate(e.target.checked)}
                className="checkbox"
                disabled={isSubmitting}
              />
              <span className="checkbox-text">Хочу участвовать в жеребьевке как игрок</span>
            </label>
          </div>

          {/* Ошибка сервера */}
          {submitError && (
            <div style={{ color: '#e74c3c', fontSize: '14px', marginBottom: '15px', textAlign: 'center' }}>
              ⚠️ {submitError}
            </div>
          )}

          {/* Кнопки */}
          <div className="game-add-buttons">
            <button 
              type="submit" 
              className="btn-primary" 
              disabled={!canSubmit}
              title={!canSubmit ? 'Заполните все обязательные поля корректно' : ''}
              style={{ opacity: isSubmitting ? 0.7 : 1 }}
            >
              {isSubmitting ? 'Создание...' : 'Создать игру'}
            </button>
            <button 
              type="button" 
              className="btn-secondary" 
              onClick={handleGoBack}
              disabled={isSubmitting}
            >
              Отмена
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default Game_add;