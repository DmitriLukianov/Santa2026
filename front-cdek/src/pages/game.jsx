import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
// Импортируем нужные методы API
import { fetchGameById, fetchParticipants, fetchMe, removeParticipant, deleteGame, finishGame, isAuthenticated } from '/src/api/gameApi.jsx';
import { addParticipant } from '/src/api/participantsApi.jsx';
import './main.css';

function Game() {
  const navigate = useNavigate();
  const { eventId } = useParams();

  // Состояния для данных
  const [gameData, setGameData] = useState(null);
  const [participantsCount, setParticipantsCount] = useState(0);
  const [isDrawDone, setIsDrawDone] = useState(false);
  const [gameStatus, setGameStatus] = useState(null);
  const [isOrganizer, setIsOrganizer] = useState(false);
  const [myParticipantId, setMyParticipantId] = useState(null);

  // Состояния загрузки и ошибок
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  // Загрузка данных при монтировании
  useEffect(() => {
    if (!isAuthenticated()) {
      navigate('/registration', { replace: true });
      return;
    }

    const loadData = async () => {
      if (!eventId) return;

      try {
        setIsLoading(true);
        setError(null);

        // 1. Параллельно загружаем игру и текущего пользователя
        const [game, me] = await Promise.all([
          fetchGameById(eventId),
          fetchMe(),
        ]);

        const isOrg = game.organizerId === me.id;
        setIsOrganizer(isOrg);

        // 2. Участников грузим отдельно — может вернуть 403 если ещё не вступил
        let participantsList = [];
        try {
          const participants = await fetchParticipants(eventId);
          participantsList = Array.isArray(participants) ? participants : (participants?.items || []);
        } catch {
          // Не участник — список недоступен, это нормально
        }

        setParticipantsCount(participantsList.length);

        // Находим participantId текущего пользователя
        const myParticipant = participantsList.find(p => p.userId === me.id);
        setMyParticipantId(myParticipant?.id || null);

        // Жеребьёвка проведена если статус игры gifting или finished
        const drawStatus = game.status === 'gifting' || game.status === 'finished';
        setIsDrawDone(drawStatus);
        setGameStatus(game.status);

        // Сохраняем основные данные игры
        setGameData({
          teamName: game.title || 'Команда',
          period: game.startDate ? new Date(game.startDate).toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' }) : '...',
          drawDate: game.drawDate ? (() => {
            const d = new Date(game.drawDate);
            const date = d.toLocaleDateString('ru-RU', { day: '2-digit', month: '2-digit', year: 'numeric' });
            const h = String(d.getHours()).padStart(2, '0');
            const m = String(d.getMinutes()).padStart(2, '0');
            return `${date} в ${h}:${m}`;
          })() : 'не указана',
          stage: game.status === 'gifting' ? 'Дарение подарков' : 'Добавление участников',
          isChatAvailable: game.status === 'gifting'
        });

      } catch (err) {
        console.error('Ошибка загрузки данных игры:', err);
        setError(err.message || 'Не удалось загрузить данные игры');
      } finally {
        setIsLoading(false);
      }
    };

    loadData();
  }, [eventId]);

  // Обработчики навигации
  const handleGoWishlist = () => {
    navigate(`/game/${eventId}/wishlist`);
  };

  const handleGoProfile = () => {
    navigate('/profile'); 
  };

  const handleGoEditGame = () => {
    navigate(`/game/${eventId}/edit`);
  };

  const handleLeaveGame = async () => {
    if (isOrganizer) {
      // Организатор удаляет игру для всех
      const confirmed = window.confirm(
        `Вы являетесь организатором игры "${gameData.teamName}".\n\nВыход удалит игру для всех участников. Это действие нельзя отменить.\n\nУдалить игру?`
      );
      if (!confirmed) return;
      try {
        await deleteGame(eventId);
        navigate('/profile');
      } catch (err) {
        alert('Не удалось удалить игру. Попробуйте позже.');
      }
    } else {
      // Обычный участник
      const warningText = isDrawDone
        ? `Жеребьёвка уже проведена.\n\nЕсли вы выйдете, ваш Санта потеряет получателя. Всё равно выйти из игры "${gameData.teamName}"?`
        : `Вы уверены, что хотите выйти из игры "${gameData.teamName}"?`;

      if (!window.confirm(warningText)) return;

      if (!myParticipantId) {
        alert('Не удалось найти вашу запись участника. Обратитесь к организатору.');
        return;
      }
      try {
        await removeParticipant(myParticipantId);
        navigate('/profile');
      } catch (err) {
        alert('Не удалось выйти из игры. Попробуйте позже.');
      }
    }
  };

  const handleJoinGame = async () => {
    try {
      const participant = await addParticipant(eventId, {});
      setMyParticipantId(participant.id);
      // Обновим счётчик участников
      setParticipantsCount(prev => prev + 1);
    } catch (err) {
      alert(err.message || 'Не удалось присоединиться к игре.');
    }
  };

  const handleFinishGame = async () => {
    const confirmed = window.confirm(
      `Завершить игру "${gameData.teamName}"?\n\nПосле завершения игра перейдёт в архив и станет недоступна для новых действий.`
    );
    if (!confirmed) return;
    try {
      await finishGame(eventId);
      setGameStatus('finished');
      setGameData(prev => ({ ...prev, stage: 'Завершена', isChatAvailable: false }));
    } catch (err) {
      alert(err.message || 'Не удалось завершить игру. Попробуйте позже.');
    }
  };

  const handleDrawResult = () => {
    if (!isDrawDone) {
      alert('Жеребьёвка ещё не проведена!');
      return;
    }
    navigate(`/game/${eventId}/letter`);
  };

  const handleSecretChat = () => {
    if (!isDrawDone) {
      alert('Секретный чат будет доступен после жеребьёвки!');
      return;
    }
    navigate(`/game/${eventId}/chat`);
  };

  // Рендер состояния загрузки
  if (isLoading) {
    return (
      <div className="overlay_game">
        <div className="card_game">
          <div style={{ textAlign: 'center', padding: '50px' }}>
            <i className="ti ti-loader" style={{ fontSize: '48px', color: '#44E858', animation: 'spin 1s linear infinite' }}></i>
            <p style={{ marginTop: '20px', color: '#757575' }}>Загрузка игры...</p>
          </div>
        </div>
      </div>
    );
  }

  // Рендер состояния ошибки
  if (error || !gameData) {
    return (
      <div className="overlay_game">
        <div className="card_game">
          <div style={{ textAlign: 'center', padding: '50px' }}>
            <i className="ti ti-alert-circle" style={{ fontSize: '48px', color: '#e74c3c' }}></i>
            <h2 style={{ marginTop: '20px', color: '#1E1E1E' }}>Ошибка</h2>
            <p style={{ color: '#757575', marginBottom: '20px' }}>{error || 'Данные не найдены'}</p>
            <button className="btn-secondary" onClick={() => navigate('/profile')}>Вернуться в профиль</button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="overlay_game">
      <div className="card_game">
        {/* Заголовок */}
        <h2 className="game-title">Тайный Санта</h2>
        
        {/* Название команды */}
        <h1 className="team-name">{gameData.teamName}</h1>

         {/* Кнопка редактирования (видна только организатору) */}
        {isOrganizer && (
          <button 
            type="button" 
            className="btn-edit-game"
            onClick={handleGoEditGame}
            title="Настройки игры"
          >
            <i 
            className="ti ti-edit" 
            style={{ 
              fontSize: '20px', 
              color: '#000000' 
            }}
          ></i>
          </button>
        )}

        {/* Информация об игре */}
        <div className="game-info">
          <p className="info-text">Начало игры: {gameData.period}</p>
          <p className="info-text">Дата жеребьёвки: {gameData.drawDate}</p>
        </div>
                  
        {/* Статус и участники - две колонки */}
        <div className="game-status-grid">
          <div className="status-box">
            <span className="status-label">Этап игры:</span>
            <span className="status-value">{gameData.stage}</span>
          </div>

          <div className="status-box">
            <span className="status-label">Количество участников:</span>
            <span className="status-value">{participantsCount}</span>
          </div>
        </div>

        {/* Кнопки - две колонки */}
        <div className="game-buttons-grid">
          <div className="buttons-column">
            <button 
              type="button" 
              className="btn-primary"  
              onClick={handleSecretChat}
              disabled={!gameData.isChatAvailable}
            >
              Секретный чат
            </button>

            <button 
              type="button" 
              className="btn-primary"
              onClick={handleDrawResult}
              disabled={!isDrawDone}
            >
              Результат жеребьёвки
            </button>
          </div>

          <div className="buttons-column">
            <button 
              type="button" 
              className="btn-primary"
              onClick={handleGoWishlist}
            >
              Проверить вишлист
            </button>
            
            <button 
              type="button" 
              className="btn-secondary"
              onClick={handleGoProfile}
            >
              Мой профиль
            </button>
          </div>
        </div>

        {/* Футер с кнопками управления */}
        <div className="game-footer">
          {isOrganizer && gameStatus === 'gifting' && (
            <button
              type="button"
              className="btn-secondary"
              onClick={handleFinishGame}
            >
              Завершить игру
            </button>
          )}
          <button
            type="button"
            className={isOrganizer ? 'btn-danger' : 'btn-secondary'}
            onClick={handleLeaveGame}
          >
            {isOrganizer ? 'Удалить игру' : 'Выйти из игры'}
          </button>
        </div>
      </div>
    </div>
  );
}

export default Game;