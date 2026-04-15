import React, { useState, useRef, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { fetchMyWishlist, fetchWishlistItems, deleteWishlistItem, updateWishlistItem, isAuthenticated, uploadFile } from '/src/api/gameApi.jsx';
import './main.css';

// === ФУНКЦИИ ВАЛИДАЦИИ ===
const validateName = (name) => {
  const errors = [];
  const trimmed = name.trim();
  if (!trimmed) {
    errors.push('Название товара обязательно');
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

const validatePrice = (price) => {
  const errors = [];
  const num = parseFloat(price);
  if (isNaN(num)) {
    errors.push('Цена должна быть числом');
    return errors;
  }
  if (num <= 0) {
    errors.push('Цена должна быть больше 0');
  }
  if (num > 1000000) {
    errors.push('Максимальная цена — 1 000 000');
  }
  const parts = price.toString().split('.');
  if (parts[1] && parts[1].length > 2) {
    errors.push('Цена может иметь не более 2 знаков после запятой');
  }
  return errors;
};

function WishlistRed() {
  const navigate = useNavigate();
  const { eventId, itemId } = useParams();
  const fileInputRef = useRef(null);

  const [formData, setFormData] = useState({
    name: '',
    price: '',
    link: ''
  });
  const [wishlistId, setWishlistId] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState(null);

  const [errors, setErrors] = useState({ name: [], price: [] });
  const [touched, setTouched] = useState({ name: false, price: false });

  const [isDragging, setIsDragging] = useState(false);
  const [files, setFiles] = useState([]);

  useEffect(() => {
    if (!isAuthenticated()) {
      navigate('/registration', { replace: true });
      return;
    }

    const loadItem = async () => {
      if (!itemId) return;

      try {
        setIsLoading(true);
        setError(null);

        // 1. Получаем вишлист (с eventId или без)
        const wishlistData = await fetchMyWishlist(eventId);
        const wId = wishlistData.id;
        setWishlistId(wId);

        // 2. Получаем товары вишлиста
        const itemsData = await fetchWishlistItems(wId);
        const itemsList = Array.isArray(itemsData) ? itemsData : (itemsData?.items || []);
        const item = itemsList.find(i => i.id === itemId);

        if (!item) {
          throw new Error('Товар не найден');
        }

        // 3. Заполняем форму
        setFormData({
          name: item.title || '',
          price: item.price ? String(item.price) : '',
          link: item.link || ''
        });

      } catch (err) {
        console.error('Ошибка загрузки товара:', err);
        setError(err.message || 'Не удалось загрузить данные товара');
      } finally {
        setIsLoading(false);
      }
    };

    loadItem();
  }, [eventId, itemId, navigate]);

  const handlePriceChange = (e) => {
    const value = e.target.value;
    if (value === '' || /^\d*\.?\d*$/.test(value)) {
      setFormData(prev => ({ ...prev, price: value }));
      if (touched.price) {
        setErrors(prev => ({ ...prev, price: validatePrice(value) }));
      }
    }
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
    if (touched[name] && name === 'name') {
      setErrors(prev => ({ ...prev, name: validateName(value) }));
    }
  };

  const handleBlur = (e) => {
    const { name, value } = e.target;
    setTouched(prev => ({ ...prev, [name]: true }));
    if (name === 'name') {
      setErrors(prev => ({ ...prev, name: validateName(value) }));
    } else if (name === 'price') {
      setErrors(prev => ({ ...prev, price: validatePrice(value) }));
    }
  };

  const isFormValid = () => {
    const nameErrors = validateName(formData.name);
    const priceErrors = validatePrice(formData.price);
    setErrors({ name: nameErrors, price: priceErrors });
    return nameErrors.length === 0 && priceErrors.length === 0;
  };

  const handleDragOver = (e) => { e.preventDefault(); setIsDragging(true); };
  const handleDragLeave = (e) => { e.preventDefault(); setIsDragging(false); };
  const handleDrop = (e) => {
    e.preventDefault();
    setIsDragging(false);
    const droppedFiles = Array.from(e.dataTransfer.files);
    setFiles(prev => [...prev, ...droppedFiles]);
  };
  const handleFileSelect = (e) => {
    const selectedFiles = Array.from(e.target.files);
    setFiles(prev => [...prev, ...selectedFiles]);
  };

  const handleGoWishlist = () => {
    if (eventId) {
      navigate(`/game/${eventId}/wishlist`);
    } else {
      navigate('/wishlist');
    }
  };

  const handleSave = async (e) => {
    e.preventDefault();

    if (!isFormValid()) {
      setTouched({ name: true, price: true });
      return;
    }

    if (!wishlistId) {
      alert('Ошибка: Вишлист не найден');
      return;
    }

    try {
      setIsSaving(true);

      // Загрузка нового файла (если выбран)
      let imageURL = '';
      if (files[0]) {
        try {
          const uploadResult = await uploadFile(files[0]);
          imageURL = uploadResult?.url || '';
        } catch (uploadErr) {
          console.warn('Не удалось загрузить изображение:', uploadErr);
        }
      }

      const itemData = {
        title: formData.name.trim(),
        price: parseFloat(formData.price),
        link: formData.link || '',
        ...(imageURL && { imageURL }),
      };

      await updateWishlistItem(wishlistId, itemId, itemData);

      handleGoWishlist();
    } catch (err) {
      console.error('Ошибка сохранения:', err);
      alert(`Не удалось сохранить: ${err.message}`);
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!window.confirm('Удалить этот товар?')) return;
    if (!wishlistId) {
      alert('Ошибка: Вишлист не найден');
      return;
    }
    try {
      await deleteWishlistItem(wishlistId, itemId);
      handleGoWishlist();
    } catch (err) {
      console.error('Ошибка удаления:', err);
      alert(`Не удалось удалить: ${err.message}`);
    }
  };

  const handleGoBack = () => {
    navigate(-1);
  };

  if (isLoading) {
    return (
      <div className="overlay_wishlist">
        <div className="card_wishlist">
          <div style={{ textAlign: 'center', padding: '40px' }}>
            <i className="ti ti-loader" style={{ fontSize: '48px', color: '#44E858', animation: 'spin 1s linear infinite' }}></i>
            <p style={{ marginTop: '20px', color: '#757575' }}>Загрузка...</p>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="overlay_wishlist">
        <div className="card_wishlist">
          <div style={{ textAlign: 'center', padding: '40px', color: '#e74c3c' }}>
            <p>{error}</p>
            <button className="btn-secondary" onClick={handleGoBack} style={{ marginTop: '20px' }}>Назад</button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="overlay_wishlist">
      <div className="card_wishlist">
        <h1>Редактирование товара</h1>

        <form onSubmit={handleSave} noValidate>
          <div className="wishlist-content">
            <div className="form-left">
              <div className="form-group">
                <label>Название товара <span className="required">*</span></label>
                <input
                  type="text"
                  name="name"
                  placeholder="Введите название"
                  value={formData.name}
                  onChange={handleChange}
                  onBlur={handleBlur}
                  required
                  disabled={isSaving}
                  className={errors.name.length > 0 && touched.name ? 'input-error' : ''}
                />
                {errors.name.length > 0 && touched.name && (
                  <ul className="error-list">
                    {errors.name.map((err, i) => (
                      <li key={i} className="error-item">• {err}</li>
                    ))}
                  </ul>
                )}
              </div>

              <div className="form-group">
                <label>Цена <span className="required">*</span></label>
                <input
                  type="text"
                  name="price"
                  placeholder="0"
                  value={formData.price}
                  onChange={handlePriceChange}
                  onBlur={handleBlur}
                  inputMode="decimal"
                  disabled={isSaving}
                  className={errors.price.length > 0 && touched.price ? 'input-error' : ''}
                />
                {errors.price.length > 0 && touched.price && (
                  <ul className="error-list">
                    {errors.price.map((err, i) => (
                      <li key={i} className="error-item">• {err}</li>
                    ))}
                  </ul>
                )}
              </div>

              <div className="form-group">
                <label>Ссылка на товар</label>
                <input
                  type="url"
                  name="link"
                  placeholder="https://..."
                  value={formData.link}
                  onChange={handleChange}
                  disabled={isSaving}
                />
              </div>
            </div>

            <div className="form-right">
              <div
                className={`upload-area ${isDragging ? 'dragover' : ''}`}
                onClick={() => fileInputRef.current?.click()}
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={handleDrop}
                style={{ opacity: isSaving ? 0.6 : 1, pointerEvents: isSaving ? 'none' : 'auto' }}
              >
                <i className="ti ti-upload" style={{ fontSize: '48px', color: '#44E858' }}></i>
                <div className="upload-text">Загрузить новый файл</div>
                <div className="upload-hint">Можно загрузить не более 1 файла</div>
                <input
                  ref={fileInputRef}
                  type="file"
                  style={{ display: 'none' }}
                  accept="image/*"
                  onChange={handleFileSelect}
                  disabled={isSaving}
                />
              </div>

              {files.length > 0 && (
                <div className="file-list">
                  <strong>Выбрано файлов: {files.length}</strong>
                  {files.map((file, index) => (
                    <div key={index} className="file-item">• {file.name}</div>
                  ))}
                </div>
              )}
            </div>
          </div>

          <div className="wishlist-red-buttons">
            <button
              type="submit"
              className="btn-primary"
              disabled={isSaving}
              style={{ opacity: isSaving ? 0.7 : 1, cursor: isSaving ? 'not-allowed' : 'pointer' }}
            >
              {isSaving ? 'Сохранение...' : 'Сохранить'}
            </button>
            <button
              type="button"
              className="btn-secondary"
              onClick={handleDelete}
              disabled={isSaving}
              style={{ borderColor: '#e74c3c', color: '#e74c3c' }}
            >
              Удалить
            </button>
            <button
              type="button"
              className="btn-secondary"
              onClick={handleGoBack}
              disabled={isSaving}
            >
              Отмена
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export default WishlistRed;
