import { useState, useEffect } from 'react';
import { AddWater, SnoozeReminder, DismissReminder, GetSettings } from '../wailsjs/go/main/App';

interface ReminderModalProps {
 consumed: number;
 goal: number;
 onClose: () => void;
}

const translations = {
 es: {
 title: "HORA DE BEBER AGUA",
 progress: "Progreso actual",
 snooze: "Posponer",
 dismiss: "Descartar",
 minutes: "min",
 },
 en: {
 title: "TIME TO DRINK WATER",
 progress: "Current progress",
 snooze: "Snooze",
 dismiss: "Dismiss",
 minutes: "min",
 },
};

function ReminderModal({ consumed, goal, onClose }: ReminderModalProps) {
 const [visible, setVisible] = useState(false);
 const [lang, setLang] = useState<'es' | 'en'>('es');
 const [currentConsumed, setCurrentConsumed] = useState(consumed);

 useEffect(() => {
 const userLang = navigator.language.startsWith('es') ? 'es' : 'en';
 setLang(userLang as 'es' | 'en');
 setTimeout(() => setVisible(true), 50);
 }, []);

 const t = translations[lang];
 const progress = Math.min(100, (currentConsumed / goal) * 100);
 const segmentsFilled = Math.min(10, Math.floor((currentConsumed / goal) * 10));

 const handleAdd = async (ml: number) => {
 try {
 await AddWater(ml);
 await DismissReminder();
 // Leer estado actualizado
 const settings = await GetSettings();
 setCurrentConsumed(settings.today_consumed);
 // Cerrar despues de mostrar el nuevo valor
 setTimeout(() => onClose(), 500);
 } catch (e) {
 console.error('Error:', e);
 }
 };

 const handleSnooze = async (minutes: number) => {
 try {
 await SnoozeReminder(minutes);
 onClose();
 } catch (e) {
 console.error('Error:', e);
 }
 };

 const handleDismiss = async () => {
 try {
 await DismissReminder();
 onClose();
 } catch (e) {
 console.error('Error:', e);
 }
 };

  return (
    <div className={`modal-overlay ${visible ? 'visible' : ''}`}>
      <div className={`modal-content ${visible ? 'visible' : ''}`}>
        {/* Header con animacion de agua */}
        <div className="modal-header">
          <div className="water-icon">
            <svg viewBox="0 0 24 24" fill="currentColor" width="48" height="48">
              <path d="M12 2C12 2 5 10 5 15C5 18.866 8.134 22 12 22C15.866 22 19 18.866 19 15C19 10 12 2 12 2ZM12 20C9.243 20 7 17.757 7 15C7 12.27 10.076 7.476 12 4.642C13.924 7.476 17 12.27 17 15C17 17.757 14.757 20 12 20Z"/>
            </svg>
          </div>
          <h2 className="modal-title">{t.title}</h2>
        </div>

        {/* Progreso actual */}
        <div className="modal-progress">
          <span className="modal-progress-label">{t.progress}</span>
          <div className="modal-progress-bar">
            <div className="modal-progress-fill" style={{ width: `${progress}%` }} />
          </div>
          <div className="modal-progress-info">
            <span>{(consumed / 1000).toFixed(1)}L / {(goal / 1000).toFixed(1)}L</span>
            <span>{Math.round(progress)}%</span>
          </div>
          <div className="modal-segments">
            {[...Array(10)].map((_, i) => (
              <div key={i} className={`modal-seg ${i < segmentsFilled ? 'on' : ''}`} />
            ))}
          </div>
        </div>

        {/* Botones de agregar agua */}
        <div className="modal-add-row">
          <button className="modal-add-btn" onClick={() => handleAdd(150)}>
            <span className="modal-add-ml">150</span>
            <span className="modal-add-unit">ml</span>
          </button>
          <button className="modal-add-btn primary" onClick={() => handleAdd(250)}>
            <span className="modal-add-ml">250</span>
            <span className="modal-add-unit">ml</span>
          </button>
          <button className="modal-add-btn" onClick={() => handleAdd(500)}>
            <span className="modal-add-ml">500</span>
            <span className="modal-add-unit">ml</span>
          </button>
        </div>

        {/* Botones de accion */}
        <div className="modal-action-row">
          <div className="snooze-options">
            <button className="snooze-btn" onClick={() => handleSnooze(5)}>5 {t.minutes}</button>
            <button className="snooze-btn" onClick={() => handleSnooze(15)}>15 {t.minutes}</button>
            <button className="snooze-btn" onClick={() => handleSnooze(30)}>30 {t.minutes}</button>
          </div>
          <button className="dismiss-btn" onClick={handleDismiss}>
            {t.dismiss}
          </button>
        </div>
      </div>
    </div>
  );
}

export default ReminderModal;
