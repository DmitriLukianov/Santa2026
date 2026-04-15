import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  fetchMyWishlist,
  fetchWishlistItems,
  addWishlistItem,
  deleteWishlistItem,
  isAuthenticated,
} from '/src/api/gameApi.jsx';
import './main.css';

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

const normalizeImageUrl = (url) => {
  if (!url) return null;
  if (url.startsWith('http://localhost:8080')) {
    return url.replace('http://localhost:8080', '');
  }
  return url;
};

function Wishlist() {
  const navigate = useNavigate();
  const { eventId: rawEventId } = useParams();
  const eventId = rawEventId && UUID_RE.test(rawEventId) ? rawEventId : undefined;

  const [gifts, setGifts] = useState([]);
  const [wishlistId, setWishlistId] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);
  const [isEmpty, setIsEmpty] = useState(true);
  const [openMenuId, setOpenMenuId] = useState(null);

  // Состояния для модалки импорта
  const [showImport, setShowImport] = useState(false);
  const [personalItems, setPersonalItems] = useState([]);
  const [selectedIds, setSelectedIds] = useState(new Set());
  const [isImporting, setIsImporting] = useState(false);
  const [isLoadingPersonal, setIsLoadingPersonal] = useState(false);

  useEffect(() => {
    if (!isAuthenticated()) {
      navigate('/registration', { replace: true });
      return;
    }

    const loadWishlist = async () => {
      try {
        setIsLoading(true);
        setError(null);

        const wishlistData = await fetchMyWishlist(eventId);
        const wId = wishlistData.id;
        setWishlistId(wId);

        const items = await fetchWishlistItems(wId);
        const list = Array.isArray(items) ? items : (items?.items || []);
        setGifts(list);
        setIsEmpty(list.length === 0);
      } catch (err) {
        console.error('Ошибка загрузки вишлиста:', err);
        // Если пользователь не участник события — показываем специальное сообщение
        const msg = err.message || '';
        if (eventId && (msg.includes('participant') || msg.includes('участник') || msg.includes('500') || msg.includes('Внутренняя'))) {
          setError('not_participant');
        } else {
          setError(msg || 'Не удалось загрузить товары');
        }
        setIsEmpty(true);
      } finally {
        setIsLoading(false);
      }
    };

    loadWishlist();
  }, [eventId, navigate]);

  const handleGoWishlist_add = () => {
    if (eventId) {
      navigate(`/game/${eventId}/wishlist/add`);
    } else {
      navigate('/wishlist/add');
    }
  };

  const handleGoWishlist_red = (id) => {
    if (eventId) {
      navigate(`/game/${eventId}/wishlist/items/${id}`);
    } else {
      navigate(`/wishlist/items/${id}`);
    }
    setOpenMenuId(null);
  };

  const handleDelete = async (itemId) => {
    if (!wishlistId) return;
    if (window.confirm('Удалить этот подарок?')) {
      try {
        const prevGifts = [...gifts];
        setGifts(prev => prev.filter(gift => gift.id !== itemId));
        setOpenMenuId(null);
        await deleteWishlistItem(wishlistId, itemId);
      } catch (err) {
        console.error('Ошибка удаления:', err);
        alert('Не удалось удалить товар. Попробуйте снова.');
        setGifts(prevGifts);
      }
    }
  };

  const handleClose = () => {
    if (eventId) {
      navigate(`/game/${eventId}`);
    } else {
      navigate('/profile');
    }
  };

  const toggleMenu = (id) => {
    setOpenMenuId(openMenuId === id ? null : id);
  };

  useEffect(() => {
    const handleClickOutside = () => setOpenMenuId(null);
    if (openMenuId) {
      document.addEventListener('click', handleClickOutside);
      return () => document.removeEventListener('click', handleClickOutside);
    }
  }, [openMenuId]);

  // Открыть модалку импорта — загрузить личный вишлист
  const handleOpenImport = async () => {
    setShowImport(true);
    setSelectedIds(new Set());
    setIsLoadingPersonal(true);
    try {
      const personalWishlist = await fetchMyWishlist(undefined); // личный (без eventId)
      const items = await fetchWishlistItems(personalWishlist.id);
      const list = Array.isArray(items) ? items : (items?.items || []);
      // Исключаем уже добавленные (по title)
      const existingTitles = new Set(gifts.map(g => g.title.toLowerCase()));
      setPersonalItems(list.filter(item => !existingTitles.has(item.title.toLowerCase())));
    } catch (err) {
      console.error('Ошибка загрузки личного вишлиста:', err);
      alert('Не удалось загрузить личный вишлист.');
      setShowImport(false);
    } finally {
      setIsLoadingPersonal(false);
    }
  };

  const toggleSelect = (id) => {
    setSelectedIds(prev => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const handleSelectAll = () => {
    if (selectedIds.size === personalItems.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(personalItems.map(i => i.id)));
    }
  };

  // Импортировать выбранные товары в вишлист события
  const handleImport = async () => {
    if (selectedIds.size === 0 || !wishlistId) return;
    setIsImporting(true);
    try {
      const toImport = personalItems.filter(item => selectedIds.has(item.id));
      const results = await Promise.allSettled(
        toImport.map(item =>
          addWishlistItem(wishlistId, {
            title: item.title,
            price: item.price,
            ...(item.link && { link: item.link }),
            ...(item.imageUrl && { imageURL: item.imageUrl }),
          })
        )
      );
      const added = results
        .filter(r => r.status === 'fulfilled')
        .map(r => r.value);
      const failed = results.filter(r => r.status === 'rejected').length;
      if (added.length > 0) {
        setGifts(prev => [...prev, ...added]);
        setIsEmpty(false);
        setShowImport(false);
      }
      if (failed > 0) {
        alert(`Не удалось добавить ${failed} из ${toImport.length} товаров.`);
      }
    } catch (err) {
      console.error('Ошибка импорта:', err);
      alert('Не удалось импортировать товары.');
    } finally {
      setIsImporting(false);
    }
  };

  return (
    <div className="overlay_wishlist">
      <div className="card_wishlist wishlist-new">

        <button className="close-wishlist" onClick={handleClose}>
          <i className="ti ti-x" style={{ fontSize: '24px', color: '#44E858' }}></i>
        </button>

        {isLoading ? (
          <>
            <div className="wishlist-header">
              <h1 className="wishlist-title">Мой вишлист</h1>
            </div>
            <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '200px' }}>
              <i className="ti ti-loader" style={{ fontSize: '36px', color: '#44E858', animation: 'spin 1s linear infinite' }}></i>
            </div>
          </>
        ) : error && gifts.length === 0 ? (
          <div className="wishlist-error">
            {error === 'not_participant' ? (
              <>
                <i className="ti ti-user-off" style={{ fontSize: '48px', color: '#e74c3c' }}></i>
                <p className="error-text">Вы не являетесь участником этой игры.</p>
                <p style={{ color: '#757575', fontSize: '14px', marginBottom: '16px' }}>
                  Сначала присоединитесь к игре, чтобы создать вишлист.
                </p>
                <button className="btn-primary" onClick={() => navigate(`/game/${eventId}`)}>
                  Перейти к игре
                </button>
              </>
            ) : (
              <>
                <i className="ti ti-alert-circle" style={{ fontSize: '48px', color: '#e74c3c' }}></i>
                <p className="error-text">{error}</p>
                <button className="btn-secondary" onClick={() => window.location.reload()}>
                  Попробовать снова
                </button>
              </>
            )}
          </div>
        ) : isEmpty ? (
          <div className="wishlist-empty">
            <div className="empty-icon">
              <i className="ti ti-gift" style={{ fontSize: '48px', color: '#44E858', animation: 'bounce 2s infinite' }}></i>
            </div>
            <h2 className="empty-title">Тут пока ничего нет</h2>
            <p className="empty-text">
              Добавьте первый товар в свой вишлист, чтобы друзья знали, что вам подарить!
            </p>
            <div style={{ display: 'flex', gap: '12px', justifyContent: 'center', flexWrap: 'wrap' }}>
              <button type="button" className="btn-primary" onClick={handleGoWishlist_add}>
                Добавить товар
              </button>
              {eventId && (
                <button type="button" className="btn-secondary" onClick={handleOpenImport}>
                  <i className="ti ti-copy" style={{ fontSize: '16px', marginRight: '6px' }}></i>
                  Импорт из личного
                </button>
              )}
            </div>
          </div>
        ) : (
          <>
            <div className="wishlist-header">
              <h1 className="wishlist-title">Мой вишлист</h1>
              <div style={{ display: 'flex', gap: '10px' }}>
                {eventId && (
                  <button type="button" className="btn-secondary" onClick={handleOpenImport}>
                    <i className="ti ti-copy" style={{ fontSize: '16px', marginRight: '6px' }}></i>
                    Импорт из личного
                  </button>
                )}
                <button type="button" className="btn-primary" onClick={handleGoWishlist_add}>
                  Добавить подарок
                </button>
              </div>
            </div>

            <div className="wishlist-scroll-container">
              <div className="wishlist-grid">
                {gifts.map((gift) => (
                  <div key={gift.id} className="gift-card">
                    <div className="gift-menu" onClick={(e) => e.stopPropagation()}>
                      <button
                        className="gift-menu-btn"
                        onClick={(e) => { e.stopPropagation(); toggleMenu(gift.id); }}
                      >
                        <i className="ti ti-dots-vertical" style={{ fontSize: '20px', color: '#757575' }}></i>
                      </button>
                      {openMenuId === gift.id && (
                        <div className="gift-menu-dropdown">
                          <button className="menu-item edit" onClick={() => handleGoWishlist_red(gift.id)}>
                            <i className="ti ti-pencil" style={{ fontSize: '16px' }}></i>
                            Редактировать
                          </button>
                          <button className="menu-item delete" onClick={() => handleDelete(gift.id)}>
                            <i className="ti ti-trash" style={{ fontSize: '16px' }}></i>
                            Удалить
                          </button>
                        </div>
                      )}
                    </div>

                    <div className="gift-content">
                      <div className="gift-image">
                        {normalizeImageUrl(gift.imageUrl) ? (
                          <img
                            src={normalizeImageUrl(gift.imageUrl)}
                            alt={gift.title}
                            onError={(e) => { e.target.style.display = 'none'; e.target.nextSibling.style.display = 'flex'; }}
                          />
                        ) : null}
                        <div style={{ display: normalizeImageUrl(gift.imageUrl) ? 'none' : 'flex', alignItems: 'center', justifyContent: 'center', width: '100%', height: '100%' }}>
                          <i className="ti ti-gift" style={{ fontSize: '40px', color: '#ccc' }}></i>
                        </div>
                      </div>
                      <div className="gift-info">
                        <h3 className="gift-name">{gift.title}</h3>
                        <p className="gift-price">
                          {gift.price ? `${Number(gift.price).toLocaleString('ru-RU')} ₽` : ''}
                        </p>
                        {gift.link && (
                          <a href={gift.link} className="gift-link" target="_blank" rel="noopener noreferrer">
                            В магазин
                            <i className="ti ti-arrow-up-right" style={{ fontSize: '14px' }}></i>
                          </a>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </>
        )}
      </div>

      {/* МОДАЛКА ИМПОРТА */}
      {showImport && (
        <div className="modal-overlay" onClick={() => !isImporting && setShowImport(false)}>
          <div className="modal-small" style={{ maxWidth: '480px', width: '90%' }} onClick={(e) => e.stopPropagation()}>
            <button className="modal-close" onClick={() => setShowImport(false)} disabled={isImporting}>×</button>
            <h3 style={{ marginBottom: '16px', fontSize: '18px' }}>Импорт из личного вишлиста</h3>

            {isLoadingPersonal ? (
              <div style={{ textAlign: 'center', padding: '24px' }}>
                <i className="ti ti-loader" style={{ fontSize: '32px', color: '#44E858', animation: 'spin 1s linear infinite' }}></i>
                <p style={{ marginTop: '8px', color: '#757575' }}>Загрузка...</p>
              </div>
            ) : personalItems.length === 0 ? (
              <p style={{ color: '#757575', textAlign: 'center', padding: '24px 0' }}>
                Нет товаров для импорта — все уже добавлены или личный вишлист пуст.
              </p>
            ) : (
              <>
                <button
                  type="button"
                  className="btn-secondary"
                  style={{ marginBottom: '12px', fontSize: '13px', padding: '6px 12px' }}
                  onClick={handleSelectAll}
                >
                  {selectedIds.size === personalItems.length ? 'Снять всё' : 'Выбрать всё'}
                </button>

                <div style={{ maxHeight: '300px', overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: '8px' }}>
                  {personalItems.map(item => (
                    <label
                      key={item.id}
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: '12px',
                        padding: '10px 12px',
                        borderRadius: '10px',
                        border: selectedIds.has(item.id) ? '2px solid #44E858' : '2px solid #eee',
                        cursor: 'pointer',
                        background: selectedIds.has(item.id) ? '#f0fff2' : '#fff',
                        transition: 'all 0.15s',
                      }}
                    >
                      <input
                        type="checkbox"
                        checked={selectedIds.has(item.id)}
                        onChange={() => toggleSelect(item.id)}
                        style={{ width: '18px', height: '18px', accentColor: '#44E858', flexShrink: 0 }}
                      />
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ fontWeight: 500, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                          {item.title}
                        </div>
                        {item.price && (
                          <div style={{ fontSize: '13px', color: '#44E858', fontWeight: 600 }}>
                            {Number(item.price).toLocaleString('ru-RU')} ₽
                          </div>
                        )}
                      </div>
                    </label>
                  ))}
                </div>

                <div style={{ display: 'flex', gap: '10px', marginTop: '16px' }}>
                  <button
                    type="button"
                    className="btn-primary"
                    onClick={handleImport}
                    disabled={selectedIds.size === 0 || isImporting}
                    style={{ flex: 1 }}
                  >
                    {isImporting ? 'Импортирование...' : `Импортировать (${selectedIds.size})`}
                  </button>
                  <button
                    type="button"
                    className="btn-secondary"
                    onClick={() => setShowImport(false)}
                    disabled={isImporting}
                  >
                    Отмена
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

export default Wishlist;
