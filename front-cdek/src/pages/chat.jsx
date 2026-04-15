import React, { useState, useRef, useEffect, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { fetchRecipientChat, fetchSenderChat, sendMessage, sendMessageToSanta, fetchAssignments, fetchGameById, fetchMe, isAuthenticated } from '/src/api/gameApi.jsx';
import './main.css';

function SecretChat() {
  const navigate = useNavigate();
  const { eventId } = useParams();

  const [activeTab, setActiveTab] = useState('recipient');
  const [message, setMessage] = useState('');

  const [messages, setMessages] = useState({ recipient: [], sender: [] });
  const [isSending, setIsSending] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [chatError, setChatError] = useState(null);
  const [myId, setMyId] = useState(null);
  const [gameTitle, setGameTitle] = useState('');
  const [chatData, setChatData] = useState({
    recipient: { title: 'Загрузка...', partner: '' },
    sender: { title: 'Загрузка...', partner: '' }
  });

  const messagesEndRef = useRef(null);
  const myIdRef = useRef(null);
  const pollingRef = useRef(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages, activeTab]);

  const formatMessages = useCallback((chatHistory, userId) => {
    if (!Array.isArray(chatHistory)) return [];
    return chatHistory.map(msg => ({
      id: msg.id,
      text: msg.content,
      sender: userId && msg.senderId === userId ? 'me' : 'them',
      time: new Date(msg.createdAt).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
    }));
  }, []);

  const loadMessages = useCallback(async (userId) => {
    if (!eventId) return;
    const uid = userId || myIdRef.current;
    try {
      const [recipientHistory, senderHistory] = await Promise.all([
        fetchRecipientChat(eventId).catch(() => []),
        fetchSenderChat(eventId).catch(() => []),
      ]);
      setMessages({
        recipient: formatMessages(recipientHistory, uid),
        sender: formatMessages(senderHistory, uid),
      });
      setChatError(null);
    } catch (err) {
      console.warn('Не удалось загрузить сообщения:', err);
      setChatError('Не удалось загрузить сообщения');
    }
  }, [eventId, formatMessages]);

  useEffect(() => {
    if (!isAuthenticated()) {
      navigate('/registration', { replace: true });
      return;
    }

    const loadData = async () => {
      if (!eventId) return;

      try {
        setIsLoading(true);

        const [me, assignments, game] = await Promise.all([
          fetchMe(),
          fetchAssignments(eventId),
          fetchGameById(eventId),
        ]);

        setGameTitle(game?.title || '');

        const userId = me?.id || null;
        setMyId(userId);
        myIdRef.current = userId;

        const assignment = Array.isArray(assignments) ? assignments[0] : null;
        const recipientName = assignment?.receiverName || 'Участник';

        setChatData({
          recipient: {
            title: `Секретный чат с ${recipientName}`,
            partner: recipientName
          },
          sender: {
            title: 'Чат с моим Сантой',
            partner: 'Тайный Санта'
          }
        });

        await loadMessages(userId);

      } catch (err) {
        console.error('Ошибка загрузки данных чата:', err);
        setChatError('Не удалось загрузить чат');
      } finally {
        setIsLoading(false);
      }
    };

    loadData();
  }, [eventId, navigate, loadMessages]);

  // Polling каждые 10 секунд
  useEffect(() => {
    if (isLoading) return;
    pollingRef.current = setInterval(() => {
      loadMessages();
    }, 10000);
    return () => clearInterval(pollingRef.current);
  }, [isLoading, loadMessages]);

  const handleSendMessage = async (e) => {
    e.preventDefault();
    if (!message.trim() || !eventId || isSending) return;

    const text = message.trim();
    const tempId = `temp_${Date.now()}`;

    const newMessage = {
      id: tempId,
      text,
      sender: 'me',
      time: new Date().toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
    };

    setMessages(prev => ({
      ...prev,
      [activeTab]: [...prev[activeTab], newMessage]
    }));
    setMessage('');
    setIsSending(true);

    try {
      if (activeTab === 'sender') {
        await sendMessageToSanta(eventId, text);
      } else {
        await sendMessage(eventId, text);
      }
      await loadMessages();
    } catch (error) {
      console.error('Ошибка отправки сообщения:', error);
      alert('Не удалось отправить сообщение. Попробуйте позже.');
      setMessages(prev => ({
        ...prev,
        [activeTab]: prev[activeTab].filter(m => m.id !== tempId)
      }));
    } finally {
      setIsSending(false);
    }
  };

  const handleGoBack = () => {
    navigate(-1);
  };

  if (isLoading) {
    return (
      <div className="chat-page">
        <div className="chat-container">
          <div style={{ textAlign: 'center', padding: '50px', width: '100%' }}>
            <i className="ti ti-loader" style={{ fontSize: '48px', color: '#44E858', animation: 'spin 1s linear infinite' }}></i>
            <p style={{ marginTop: '20px', color: '#757575' }}>Загрузка чата...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="chat-page">
      <div className="chat-container">
        <button className="close" onClick={handleGoBack}>
          <i className="ti ti-x" style={{ fontSize: '24px', color: '#000000' }}></i>
        </button>

        <div className="chat-tabs">
          <button
            className={`chat-tab ${activeTab === 'recipient' ? 'active' : ''}`}
            onClick={() => setActiveTab('recipient')}
          >
            <i className="ti ti-gift" style={{ fontSize: '24px', color: '#000000' }}></i> Тому, кому дарю
          </button>
          <button
            className={`chat-tab ${activeTab === 'sender' ? 'active' : ''}`}
            onClick={() => setActiveTab('sender')}
          >
            <i className="ti ti-christmas-tree" style={{ fontSize: '24px', color: '#000000' }}></i> Кто дарит мне
          </button>
        </div>

        <div className="chat-header">
          <h1 className="chat-title">{chatData[activeTab].title}</h1>
          {gameTitle && <h2 className="chat-team">{gameTitle}</h2>}
          <p className="chat-partner">Собеседник: {chatData[activeTab].partner}</p>
        </div>

        <div className="chat-messages">
          {chatError ? (
            <div style={{ textAlign: 'center', marginTop: '40px' }}>
              <i className="ti ti-alert-circle" style={{ fontSize: '32px', color: '#e74c3c' }}></i>
              <p style={{ color: '#e74c3c', marginTop: '8px' }}>{chatError}</p>
              <button className="btn-secondary" style={{ marginTop: '12px' }} onClick={() => loadMessages()}>
                Попробовать снова
              </button>
            </div>
          ) : messages[activeTab].length === 0 ? (
            <div style={{ textAlign: 'center', color: '#757575', marginTop: '40px' }}>
              Пока нет сообщений. Напишите первым!
            </div>
          ) : (
            messages[activeTab].map((msg) => (
              <div
                key={msg.id}
                className={`message ${msg.sender === 'me' ? 'message-me' : 'message-them'}`}
              >
                {msg.sender === 'them' && (
                  <div className="message-avatar">
                    <i className="ti ti-gift" style={{ fontSize: '24px', color: '#44E858' }}></i>
                  </div>
                )}
                <div className="message-bubble">
                  <p className="message-text">{msg.text}</p>
                  <span className="message-time">{msg.time}</span>
                </div>
                {msg.sender === 'me' && (
                  <div className="message-avatar me">
                    <i className="ti ti-user"></i>
                  </div>
                )}
              </div>
            ))
          )}
          <div ref={messagesEndRef} />
        </div>

        <form className="chat-input-form" onSubmit={handleSendMessage}>
          <input
            type="text"
            className="chat-input"
            placeholder="Введите текст..."
            value={message}
            onChange={(e) => setMessage(e.target.value)}
          />
          <button type="submit" className="chat-send-btn" disabled={!message.trim() || isSending}>
            ➤
          </button>
        </form>
      </div>
    </div>
  );
}

export default SecretChat;
