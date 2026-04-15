import React, { useState, useEffect } from 'react';
import { useNavigate, useParams, useLocation } from 'react-router-dom';
import { fetchParticipantWishlist, fetchWishlistItems, isAuthenticated } from '/src/api/gameApi.jsx';
import './main.css';

function Wishlist_Santa({ participantName: propName }) {
  const navigate = useNavigate();
  const { eventId, participantSlug } = useParams();
  const { search } = useLocation();

  const queryName = new URLSearchParams(search).get('name');
  const displayName = propName || queryName || 'Участник';

  const [gifts, setGifts] = useState([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isEmpty, setIsEmpty] = useState(false);

  useEffect(() => {
    if (!isAuthenticated()) {
      navigate('/registration', { replace: true });
      return;
    }

    const loadWishlist = async () => {
      if (!eventId || !participantSlug) {
        setIsLoading(false);
        setIsEmpty(true);
        return;
      }

      try {
        setIsLoading(true);

        // 1. Получаем вишлист участника (participantSlug — это participantId)
        const wishlistData = await fetchParticipantWishlist(participantSlug, eventId);
        const wishlistId = wishlistData?.id;

        if (!wishlistId) {
          setIsEmpty(true);
          return;
        }

        // 2. Получаем товары вишлиста (передаём eventId для проверки Santa-доступа)
        const items = await fetchWishlistItems(wishlistId, eventId);
        const list = Array.isArray(items) ? items : (items?.items || []);
        setGifts(list);
        setIsEmpty(list.length === 0);

      } catch (err) {
        console.error('Ошибка загрузки вишлиста участника:', err);
        // Если вишлист не найден — просто показываем пустой экран
        setIsEmpty(true);
      } finally {
        setIsLoading(false);
      }
    };

    loadWishlist();
  }, [eventId, participantSlug, navigate]);

  const handleClose = () => {
    navigate(-1);
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
              <h1 className="wishlist-title">Вишлист</h1>
            </div>
            <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '200px' }}>
              <i className="ti ti-loader" style={{ fontSize: '36px', color: '#44E858', animation: 'spin 1s linear infinite' }}></i>
            </div>
          </>
        ) : isEmpty ? (
          <div className="wishlist-empty">
            <div className="empty-icon">
              <i className="ti ti-gift" style={{ fontSize: '48px', color: '#44E858', animation: 'bounce 2s infinite' }}></i>
            </div>
            <h2 className="empty-title">Тут пока ничего нет</h2>
            <p className="empty-text">
              Подождите, пока {displayName} добавит товары, или напишите в секретный чат!
            </p>
          </div>
        ) : (
          <>
            <div className="wishlist-header">
              <h1 className="wishlist-title">Вишлист: {displayName}</h1>
            </div>

            <div className="wishlist-scroll-container">
              <div className="wishlist-grid">
                {gifts.map((gift) => (
                  <div key={gift.id} className="gift-card">
                    <div className="gift-content">
                      <div className="gift-image">
                        <img
                          src={gift.imageUrl || '/placeholder.png'}
                          alt={gift.title}
                          onError={(e) => { e.target.src = '/placeholder.png'; }}
                        />
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
    </div>
  );
}

export default Wishlist_Santa;
