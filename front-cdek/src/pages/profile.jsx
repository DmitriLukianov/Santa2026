import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
// Импортируем нужные методы API
import { fetchMe, fetchUserGames, deleteGame, joinGameByLink, logout as apiLogout, isAuthenticated } from '/src/api/gameApi.jsx';
import './main.css';

function Profile() {
  const navigate = useNavigate();

  // Состояния для данных
  const [user, setUser] = useState(null);
  const [games, setGames] = useState([]);
  
  // Состояния загрузки и ошибок
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  // Состояния для модального окна подключения
  const [isJoinModalOpen, setIsJoinModalOpen] = useState(false);
  const [joinLink, setJoinLink] = useState('');
  const [joinError, setJoinError] = useState('');
  const [isJoining, setIsJoining] = useState(false);

  // ← НОВОЕ: Загрузка данных профиля и игр
  useEffect(() => {
    if (!isAuthenticated()) {
      navigate('/registration', { replace: true });
      return;
    }

    const loadData = async () => {
      try {
        setIsLoading(true);
        setError(null);

        // Параллельная загрузка профиля и списка игр
        const [userData, gamesData] = await Promise.all([
          fetchMe(),
          fetchUserGames()
        ]);

        setUser(userData);
        
        // Адаптируем ответ API под наш список (обычно это массив или объект { items: [] })
        const gamesList = Array.isArray(gamesData) ? gamesData : (gamesData.items || gamesData.data || []);
        setGames(gamesList);

      } catch (err) {
        console.error('Ошибка загрузки данных:', err);
        setError(err.message || 'Не удалось загрузить данные профиля');
        
        // Если ошибка 401 (Unauthorized), можно сразу редиректить на вход
        if (err.message.includes('401')) {
          apiLogout();
          navigate('/registration', { replace: true });
        }
      } finally {
        setIsLoading(false);
      }
    };

    loadData();
  }, [navigate]);

  const handleGoHome = () => {
    navigate('/'); 
  };

  const handleGoWishlist = (eventId) => {
    // Переходим к вишлисту конкретной игры (или общей странице выбора)
    if (eventId) {
      navigate(`/game/${eventId}/wishlist`);
    } else {
      navigate('/wishlist'); 
    }
  };

  const handleGoProfileRed = () => {
    navigate('/profile/edit');
  };

  const handleGoGameAdd = () => {
    navigate('/game/add');
  };

  const handleGoGame = (eventId) => {
    navigate(`/game/${eventId}`);
  };

  // Открытие модалки
  const openJoinModal = () => {
    setIsJoinModalOpen(true);
    setJoinLink('');
    setJoinError('');
  };

  // Закрытие модалки
  const closeJoinModal = () => {
    setIsJoinModalOpen(false);
    setJoinLink('');
    setJoinError('');
  };

  // Обработка подключения к игре
  const handleJoinSubmit = async () => {
    if (!joinLink.trim()) {
      setJoinError('Введите ссылку-приглашение');
      return;
    }

    setIsJoining(true);
    setJoinError('');

    try {
      const data = await joinGameByLink(joinLink.trim());
      
      // Извлекаем ID игры из ответа (адаптируйте под структуру вашего API)
      const gameId = data?.eventId ?? data?.event?.id ?? data?.id;
      
      closeJoinModal();
      
      if (gameId) {
        navigate(`/game/${gameId}`);
      } else {
        // Если ID нет в ответе, просто обновляем список игр
        const updatedGames = await fetchUserGames();
        setGames(Array.isArray(updatedGames) ? updatedGames : (updatedGames.items || []));
      }
    } catch (err) {
      setJoinError(err.message || 'Не удалось подключиться к игре');
    } finally {
      setIsJoining(false);
    }
  };

  const handleDeleteGame = async (game) => {
    const name = game.title || `Игра #${game.id}`;
    if (!window.confirm(`Удалить игру "${name}"?\nЭто действие нельзя отменить.`)) return;
    try {
      await deleteGame(game.id);
      setGames(prev => prev.filter(g => g.id !== game.id));
    } catch (err) {
      alert('Не удалось удалить игру. Попробуйте позже.');
    }
  };

  const handleLogout = () => {
    if (window.confirm('Вы действительно хотите выйти из профиля?')) {
      apiLogout(); // Очищаем localStorage
      navigate('/', { replace: true });
    }
  };

  // Рендер состояния загрузки
  if (isLoading) {
    return (
      <div className="overlay_profile">
        <div className="card_profile">
          <div style={{ textAlign: 'center', padding: '50px' }}>
            <i className="ti ti-loader" style={{ fontSize: '48px', color: '#44E858', animation: 'spin 1s linear infinite' }}></i>
            <p style={{ marginTop: '20px', color: '#757575' }}>Загрузка профиля...</p>
          </div>
        </div>
      </div>
    );
  }

  // Рендер состояния ошибки
  if (error && !user) {
    return (
      <div className="overlay_profile">
        <div className="card_profile">
          <div style={{ textAlign: 'center', padding: '50px' }}>
            <i className="ti ti-alert-circle" style={{ fontSize: '48px', color: '#e74c3c' }}></i>
            <h2 style={{ marginTop: '20px', color: '#1E1E1E' }}>Ошибка</h2>
            <p style={{ color: '#757575', marginBottom: '20px' }}>{error}</p>
            <button className="btn-primary" onClick={() => window.location.reload()}>Попробовать снова</button>
            <button className="btn-secondary" onClick={handleGoHome} style={{ marginLeft: '10px' }}>На главную</button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="overlay_profile">
      <div className="card_profile">

        <button 
          type="button" 
          className="btn-logout" 
          onClick={handleLogout}
          title="Выйти из профиля"
        >
          <i className="ti ti-logout" style={{ fontSize: '20px' }}></i>
        </button>
        
        <h1>МОЙ ПРОФИЛЬ</h1>

        <div className="profile-content">
          {/* Левая колонка: Личные данные */}
          <div className="column personal-data">
            <h3>ЛИЧНЫЕ ДАННЫЕ</h3>
            <form>
              <div className="input-readonly">
                <span className="value">{user?.name || 'Имя не указано'}</span>
              </div>

              <div className="input-readonly">
                <span className="value">{user?.email || 'Email не указан'}</span>
              </div>

              <button type="button" className="btn-secondary" onClick={handleGoProfileRed}>
                Изменить данные
              </button>

              <button type="button" className="btn-primary" onClick={() => navigate('/wishlist')}>
                Мой вишлист
              </button>

              <button type="button" className="btn-secondary" onClick={handleGoHome}>
                На главную
              </button>
            </form>
          </div>

          {/* Правая колонка: Игры */}
          <div className="column games-list">
            <h3>МОИ ИГРЫ ({games.length})</h3>
            
            <div className="games-scroll-container">
              {games.length === 0 ? (
                <p style={{ textAlign: 'center', color: '#757575', padding: '20px' }}>
                  У вас пока нет активных игр. <br/> Создайте новую или подключитесь к существующей!
                </p>
              ) : (
                games.map((game) => (
                  <div key={game.id} className="game-list-row">
                    <button
                      type="button"
                      className="game-item"
                      onClick={() => handleGoGame(game.id)}
                    >
                      <span className="game-item-title">
                        {(() => { const t = game.title || `Игра #${game.id}`; return t.length > 30 ? t.slice(0, 30) + '...' : t; })()}
                      </span>
                      {game.status && (
                        <span
                          className="game-item-status"
                          style={{ color: game.status === 'gifting' ? '#44E858' : game.status === 'finished' ? '#aaaaaa' : '#757575' }}
                        >
                          ● {game.status === 'gifting' ? 'Дарение' : game.status === 'finished' ? 'Завершена' : 'Регистрация'}
                        </span>
                      )}
                    </button>
                    {game.organizerId === user?.id && (
                      <button
                        type="button"
                        className="game-delete-btn"
                        onClick={() => handleDeleteGame(game)}
                        title="Удалить игру"
                      >
                        <i className="ti ti-trash" style={{ fontSize: '18px', color: '#e74c3c' }}></i>
                      </button>
                    )}
                  </div>
                ))
              )}
            </div>

            {/* Блок кнопок управления играми */}
            <div className="game-actions" style={{ display: 'flex', flexDirection: 'column', gap: '10px', marginTop: '15px' }}>
              <button type="button" className="btn-primary" onClick={handleGoGameAdd}>
                + Создать новую игру
              </button>
              
              <button type="button" className="btn-secondary" onClick={openJoinModal}>
                Подключиться к игре
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* === МОДАЛЬНОЕ ОКНО ПОДКЛЮЧЕНИЯ === */}
      {isJoinModalOpen && (
        <div className="modal-overlay" onClick={closeJoinModal}>
          <div className="modal-small" onClick={(e) => e.stopPropagation()}>
            <button className="modal-close" onClick={closeJoinModal}>×</button>
            
            <h3>Подключение к игре</h3>
            <p className="modal-label">Вставьте ссылку-приглашение:</p>
            
            <div className="link-row" style={{ flexDirection: 'column', alignItems: 'stretch' }}>
              <input 
                type="text" 
                className={`link-input ${joinError ? 'input-error' : ''}`}
                placeholder="https://site.com/join/..."
                value={joinLink}
                onChange={(e) => {
                  setJoinLink(e.target.value);
                  if (joinError) setJoinError('');
                }}
                disabled={isJoining}
                style={{ marginBottom: '10px' }}
                onKeyDown={(e) => e.key === 'Enter' && handleJoinSubmit()}
              />
              
              <button 
                type="button" 
                className="btn-primary"
                onClick={handleJoinSubmit}
                disabled={isJoining}
                style={{ width: '100%', opacity: isJoining ? 0.7 : 1 }}
              >
                {isJoining ? 'Подключение...' : 'Войти в игру'}
              </button>
            </div>

            {joinError && <span className="error-text" style={{ marginTop: '8px', display: 'block', color: '#e74c3c' }}>{joinError}</span>}
          </div>
        </div>
      )}

    </div>
  );
}

export default Profile;