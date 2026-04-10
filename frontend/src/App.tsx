import { useState, useEffect } from 'react';
import { Events } from '@wailsio/runtime';
import { 
 GetSettings, AddWater, SetWeight, SetHeight, SetLanguage, SetLocation, 
 SetReminderInterval, GetWeeklyData, GetMonthlyData, SetMood, 
 GetMoodRecommendation 
} from '../bindings/hydrapotion/app';
import { Settings, HistoryDay, MoodEntry, Mood } from '../bindings/hydrapotion/models';
import ReminderModal from './ReminderModal';
import './App.css';

interface MoodRecommendation {
 mood: string;
 base_goal: number;
 adjusted_goal: number;
 recommendation: string;
 adjustment: string;
 multiplier: number;
}

const translations = {
  es: {
    app_name: "Hydrapotion",
    daily_progress: "Progreso Diario",
    register_intake: "Registrar Intake",
    mood: "Estado Mental",
    history: "Historial",
    weather: "Clima",
    settings: "Ajustes",
    week: "Semana",
    month: "Mes",
    avg: "Promedio",
    total: "Total",
    placeholder: "Buscar ciudad...",
    mood_well: "Bien",
    mood_neutral: "Neutral",
    mood_low: "Bajo",
    mood_tense: "Tenso",
    remaining: "faltan",
    completed: "completado",
    next_reminder: "Proximo recordatorio",
    timer_active: "Timer activo",
 weight: "Peso",
 height: "Estatura",
 cm: "cm",
 kg: "kg",
    reminder_interval: "Intervalo",
    minutes: "minutos",
    save: "Guardar",
    location_label: "Ubicacion",
    streak: "Racha",
    days: "dias",
    best_day: "Mejor dia",
    goal_achieved: "Meta alcanzada",
    times: "veces",
  },
  en: {
    app_name: "Hydrapotion",
    daily_progress: "Daily Progress",
    register_intake: "Log Intake",
    mood: "Mental State",
    history: "History",
    weather: "Weather",
    settings: "Settings",
    week: "Week",
    month: "Month",
    avg: "Average",
    total: "Total",
    placeholder: "Search city...",
    mood_well: "Well",
    mood_neutral: "Neutral",
    mood_low: "Low",
    mood_tense: "Tense",
    remaining: "remaining",
    completed: "completed",
    next_reminder: "Next reminder",
    timer_active: "Timer active",
 weight: "Weight",
 height: "Height",
 cm: "cm",
 kg: "kg",
    reminder_interval: "Interval",
    minutes: "minutes",
    save: "Save",
    location_label: "Location",
    streak: "Streak",
    days: "days",
    best_day: "Best day",
    goal_achieved: "Goal achieved",
    times: "times",
  },
};

const moodEmojis: Record<string, { emoji: string; color: string }> = {
  well: { emoji: "😊", color: "#4ade80" },
  neutral: { emoji: "😐", color: "#94a3b8" },
  low: { emoji: "😔", color: "#60a5fa" },
  tense: { emoji: "😤", color: "#f87171" },
};

function App() {
  const [theme, setTheme] = useState<'dark' | 'light'>('dark');
  const [activeNav, setActiveNav] = useState('home');
 const [settings, setSettings] = useState<Settings>({
 weight: 70,
 height: 170,
 today_consumed: 0,
 daily_goal: 2450,
 language: "es",
 location: "",
 last_reset_date: "",
 reminder_interval: 1800,
 current_mood: "neutral",
 });
  const [historyData, setHistoryData] = useState<HistoryDay[]>([]);
  const [period, setPeriod] = useState<"week" | "month">("week");
  const [location, setLocation] = useState("");
  const [weather, setWeather] = useState<{ temp: number; desc: string; hydroRec: string } | null>(null);
  const [moodRec, setMoodRec] = useState<MoodRecommendation | null>(null);
  const [showReminder, setShowReminder] = useState(false);
  const [reminderData, setReminderData] = useState<{ consumed: number; goal: number } | null>(null);
  const [reminderTimer, setReminderTimer] = useState(0);
  const [timerPercent, setTimerPercent] = useState(100);
 const [weightInput, setWeightInput] = useState("70");
 const [heightInput, setHeightInput] = useState("170");
 const [intervalInput, setIntervalInput] = useState("30");

  const t = translations[settings.language as "es" | "en"] || translations.es;
  const consumed = settings.today_consumed;
  const goal = settings.daily_goal;
  const progress = Math.min(100, (consumed / goal) * 100);
  const remaining = Math.max(0, goal - consumed);
  const segmentsFilled = Math.min(10, Math.floor((consumed / goal) * 10));

  // Timer countdown effect
  useEffect(() => {
    let interval: number;
    
    const startTimer = () => {
      const startTime = Date.now();
      const endTime = startTime + (settings.reminder_interval * 1000);
      
      interval = setInterval(() => {
        const now = Date.now();
        const remainingMs = endTime - now;
        
        if (remainingMs <= 0) {
          clearInterval(interval);
          setReminderTimer(0);
          setTimerPercent(0);
        } else {
          setReminderTimer(Math.ceil(remainingMs / 1000));
          setTimerPercent((remainingMs / (settings.reminder_interval * 1000)) * 100);
        }
      }, 1000);
    };

    startTimer();

    return () => {
      if (interval) clearInterval(interval);
    };
  }, [settings.reminder_interval]);

  useEffect(() => {
    loadSettings();
    loadHistory();

    Events.On('show-reminder', (data: { consumed: number; goal: number }) => {
      setReminderData(data);
      setShowReminder(true);
    });
  }, []);

 useEffect(() => {
 if (settings.location && !location) {
 setLocation(settings.location);
 }
 setWeightInput(settings.weight.toString());
 setHeightInput((settings.height || 170).toString());
 setIntervalInput(Math.floor(settings.reminder_interval / 60).toString());
 }, [settings]);

  useEffect(() => {
    if (settings.location) {
      fetchWeather(settings.location);
    }
  }, [settings.location]);

  useEffect(() => {
    if (settings.current_mood && weather) {
      loadMoodRecommendation();
    }
  }, [settings.current_mood, weather?.temp]);

  const loadMoodRecommendation = async () => {
    try {
      const temp = weather?.temp || 20;
      const rec = await GetMoodRecommendation(temp);
      setMoodRec(rec);
    } catch (e) {
      console.error("Failed to load mood recommendation:", e);
    }
  };

  const fetchWeather = async (city: string) => {
    try {
      const geoRes = await fetch(`https://geocoding-api.open-meteo.com/v1/search?name=${encodeURIComponent(city)}&count=1`);
      const geoData = await geoRes.json();

      if (!geoData.results || geoData.results.length === 0) return;

      const { latitude, longitude } = geoData.results[0];
      const weatherRes = await fetch(
        `https://api.open-meteo.com/v1/forecast?latitude=${latitude}&longitude=${longitude}&current_weather=true`
      );
      const weatherData = await weatherRes.json();

      const temp = Math.round(weatherData.current_weather.temperature);
      const wmoCode = weatherData.current_weather.weathercode;
      const desc = getWeatherDescription(wmoCode, settings.language);
      const hydroRec = getHydroRecommendation(temp, settings.language);

      setWeather({ temp, desc, hydroRec });
    } catch (e) {
      console.error("Failed to fetch weather:", e);
    }
  };

  const getWeatherDescription = (code: number, lang: string): string => {
    const descriptions: Record<number, { es: string; en: string }> = {
      0: { es: "Despejado", en: "Clear" },
      1: { es: "Mayormente despejado", en: "Mostly clear" },
      2: { es: "Parcialmente nublado", en: "Partly cloudy" },
      3: { es: "Nublado", en: "Overcast" },
      45: { es: "Neblina", en: "Fog" },
      48: { es: "Neblina con escarcha", en: "Depositing rime fog" },
      51: { es: "Llovizna ligera", en: "Light drizzle" },
      53: { es: "Llovizna moderada", en: "Moderate drizzle" },
      55: { es: "Llovizna intensa", en: "Dense drizzle" },
      61: { es: "Lluvia ligera", en: "Slight rain" },
      63: { es: "Lluvia moderada", en: "Moderate rain" },
      65: { es: "Lluvia intensa", en: "Heavy rain" },
      71: { es: "Nevada ligera", en: "Slight snow" },
      73: { es: "Nevada moderada", en: "Moderate snow" },
      75: { es: "Nevada intensa", en: "Heavy snow" },
      80: { es: "Chubascos ligeros", en: "Slight showers" },
      81: { es: "Chubascos moderados", en: "Moderate showers" },
      82: { es: "Chubascos violentos", en: "Violent showers" },
      95: { es: "Tormenta", en: "Thunderstorm" },
    };
    return descriptions[code]?.[lang] || (lang === "es" ? "Variable" : "Variable");
  };

  const getHydroRecommendation = (temp: number, lang: string): string => {
    if (temp >= 30) return lang === "es" ? "Bebe mas agua de lo habitual" : "Drink more water than usual";
    if (temp >= 25) return lang === "es" ? "Mantente hidratado" : "Stay hydrated";
    if (temp <= 10) return lang === "es" ? "Agua templada ayuda" : "Warm water helps";
    return "";
  };

  useEffect(() => {
    loadHistory();
  }, [period]);

  const loadSettings = async () => {
    try {
      const s = await GetSettings();
      setSettings(s);
    } catch (e) {
      console.error("Failed to load settings:", e);
    }
  };

  const loadHistory = async () => {
    try {
      const data = period === "week" ? await GetWeeklyData() : await GetMonthlyData();
      setHistoryData(data);
    } catch (e) {
      console.error("Failed to load history:", e);
    }
  };

  const playSound = () => {
    try {
      const audio = new Audio('/sound.wav');
      audio.volume = 0.5;
      audio.play().catch(() => {});
    } catch {}
  };

  const addWater = async (ml: number) => {
    try {
      const s = await AddWater(ml);
      setSettings(s);
      loadHistory();
      playSound();
    } catch (e) {
      console.error("Failed to add water:", e);
    }
  };

  const setMoodHandler = async (mood: string) => {
    try {
      const s = await SetMood(mood);
      setSettings(s);
    } catch (e) {
      console.error("Failed to set mood:", e);
    }
  };

 const setWeightHandler = async () => {
 const weight = parseInt(weightInput);
 if (weight > 0) {
 try {
 const s = await SetWeight(weight);
 setSettings(s);
 } catch (e) {
 console.error("Failed to set weight:", e);
 }
 }
 };

 const setHeightHandler = async () => {
 const height = parseInt(heightInput);
 if (height > 0) {
 try {
 const s = await SetHeight(height);
 setSettings(s);
 } catch (e) {
 console.error("Failed to set height:", e);
 }
 }
 };

 const setLanguage = async (language: string) => {
    try {
      const s = await SetLanguage(language);
      setSettings(s);
    } catch (e) {
      console.error("Failed to set language:", e);
    }
  };

  const setReminderHandler = async () => {
    const interval = parseInt(intervalInput) * 60;
    if (interval > 0) {
      try {
        const s = await SetReminderInterval(interval);
        setSettings(s);
      } catch (e) {
        console.error("Failed to set reminder:", e);
      }
    }
  };

  const setLocationHandler = async () => {
    try {
      const s = await SetLocation(location);
      setSettings(s);
      if (location) fetchWeather(location);
    } catch (e) {
      console.error("Failed to set location:", e);
    }
  };

  const toggleTheme = () => {
    setTheme(theme === 'dark' ? 'light' : 'dark');
  };

  const toggleLanguage = () => {
    setLanguage(settings.language === 'es' ? 'en' : 'es');
  };

  const formatTimer = (seconds: number): string => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const maxMl = Math.max(...historyData.map(d => d.ml), goal);
  const totalMl = historyData.reduce((sum, d) => sum + d.ml, 0);
  const avgMl = historyData.length > 0 ? Math.round(totalMl / historyData.length) : 0;
  const goalsMet = historyData.filter(d => d.ml >= goal).length;
  const bestDay = historyData.length > 0 ? Math.max(...historyData.map(d => d.ml)) : 0;

  return (
    <div className="app-container" data-theme={theme}>
 {/* Sidebar */}
 <aside className="sidebar">
 <div className="sidebar-logo">
 <img src="/logo-bgremoved.png" alt="Hydrapotion" />
 </div>
        
        <nav className="sidebar-nav">
          <button 
            className={`nav-btn ${activeNav === 'home' ? 'active' : ''}`}
            onClick={() => setActiveNav('home')}
            title={t.app_name}
          >
            <span className="material-icons-outlined">home</span>
          </button>
          <button 
            className={`nav-btn ${activeNav === 'history' ? 'active' : ''}`}
            onClick={() => setActiveNav('history')}
            title={t.history}
          >
            <span className="material-icons-outlined">bar_chart</span>
          </button>
          <button 
            className={`nav-btn ${activeNav === 'settings' ? 'active' : ''}`}
            onClick={() => setActiveNav('settings')}
            title={t.settings}
          >
            <span className="material-icons-outlined">settings</span>
          </button>
        </nav>

        <div className="sidebar-controls">
          <button className="control-btn lang-btn" onClick={toggleLanguage}>
            {settings.language === 'es' ? 'ES' : 'EN'}
          </button>
          <button className="control-btn" onClick={toggleTheme}>
            <span className="material-icons-outlined" style={{ fontSize: '18px' }}>
              {theme === 'dark' ? 'light_mode' : 'dark_mode'}
            </span>
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className={`main-content ${activeNav}`}>
        
        {/* ==================== HOME VIEW ==================== */}
        {activeNav === 'home' && (
          <div className="content-grid home-grid">
            {/* Progress Card */}
            <div className="card progress-card">
              <div className="card-header">
                <span className="card-label">
                  <span className="material-icons-outlined">water_drop</span>
                  {t.daily_progress}
                </span>
              </div>
              
              <div className="progress-main">
                <div className="progress-value">
                  <span className="progress-number">{(consumed / 1000).toFixed(1)}</span>
                  <span className="progress-unit">L</span>
                </div>
                <span className="progress-goal">meta {(goal / 1000).toFixed(1)} L</span>
              </div>
              
              <div className="progress-bar-track">
                <div className="progress-bar-fill" style={{ width: `${progress}%` }} />
              </div>
              
              <div className="progress-info">
                <span>{Math.round(progress)}% {t.completed}</span>
                <span>{t.remaining} {(remaining / 1000).toFixed(1)} L</span>
              </div>
              
              <div className="progress-segments">
                {[...Array(10)].map((_, i) => (
                  <div key={i} className={`segment ${i < segmentsFilled ? 'filled' : ''}`} />
                ))}
              </div>
            </div>

            {/* Mood Card */}
            <div className="card mood-card">
              <div className="card-header">
                <span className="card-label">
                  <span className="material-icons-outlined">mood</span>
                  {t.mood}
                </span>
              </div>
              
              <div className="mood-grid">
                {Object.entries(moodEmojis).map(([mood, data]) => (
                  <button
                    key={mood}
                    className={`mood-btn ${settings.current_mood === mood ? 'active' : ''}`}
                    onClick={() => setMoodHandler(mood)}
                  >
                    <span className="mood-emoji">{data.emoji}</span>
                    <span className="mood-label">{t[`mood_${mood}` as keyof typeof t]}</span>
                  </button>
                ))}
              </div>
              
              {moodRec && (
                <div className="mood-recommendation">
                  <span className="mood-rec-text">{moodRec.recommendation}</span>
                  {moodRec.multiplier > 1 && (
                    <div className="mood-rec-adjust">{moodRec.adjustment}</div>
                  )}
                </div>
              )}
            </div>

            {/* Intake Card */}
            <div className="card intake-card">
              <div className="card-header">
                <span className="card-label">
                  <span className="material-icons-outlined">add_circle</span>
                  {t.register_intake}
                </span>
              </div>
              
              <div className="intake-grid">
                <button className="intake-btn" onClick={() => addWater(100)}>
                  <span className="intake-amount">100</span>
                  <span className="intake-unit">ml</span>
                </button>
                <button className="intake-btn" onClick={() => addWater(150)}>
                  <span className="intake-amount">150</span>
                  <span className="intake-unit">ml</span>
                </button>
                <button className="intake-btn primary" onClick={() => addWater(250)}>
                  <span className="intake-amount">250</span>
                  <span className="intake-unit">ml</span>
                </button>
                <button className="intake-btn" onClick={() => addWater(500)}>
                  <span className="intake-amount">500</span>
                  <span className="intake-unit">ml</span>
                </button>
              </div>
            </div>

            {/* Timer Card */}
            <div className="card timer-card">
              <div className="card-header">
                <span className="card-label">
                  <span className="material-icons-outlined">timer</span>
                  {t.next_reminder}
                </span>
              </div>
              
              <div className="timer-display">
                <div className={`timer-ring-container ${timerPercent < 20 ? 'warning' : ''}`}>
                  <svg className="timer-ring-bg" viewBox="0 0 100 100">
                    <circle
                      cx="50"
                      cy="50"
                      r="45"
                      fill="none"
                      stroke="rgba(0, 229, 255, 0.1)"
                      strokeWidth="6"
                    />
                  </svg>
                  <svg 
                    className="timer-ring-bg" 
                    viewBox="0 0 100 100"
                    style={{ 
                      transform: 'rotate(-90deg)',
                      position: 'absolute',
                      top: 0,
                      left: 0
                    }}
                  >
                    <circle
                      cx="50"
                      cy="50"
                      r="45"
                      fill="none"
                      stroke={timerPercent < 20 ? '#f59e0b' : '#00E5FF'}
                      strokeWidth="6"
                      strokeLinecap="round"
                      strokeDasharray={`${2 * Math.PI * 45}`}
                      strokeDashoffset={`${2 * Math.PI * 45 * (1 - timerPercent / 100)}`}
                      style={{ 
                        filter: timerPercent < 20 
                          ? 'drop-shadow(0 0 8px rgba(245, 158, 11, 0.6))' 
                          : 'drop-shadow(0 0 8px rgba(0, 229, 255, 0.6))',
                        transition: 'stroke-dashoffset 0.5s ease'
                      }}
                    />
                  </svg>
                  <div className="timer-inner">
                    <span className="timer-time">{formatTimer(reminderTimer)}</span>
                    <div className="timer-unit">{t.timer_active}</div>
                  </div>
                </div>
              </div>
              
              <div className={`timer-status ${timerPercent < 20 ? 'warning' : ''}`}>
                <span className="material-icons-outlined">
                  {timerPercent < 20 ? 'alarm' : 'schedule'}
                </span>
                {timerPercent < 20 
                  ? (settings.language === 'es' ? 'Bebe pronto!' : 'Drink soon!')
                  : (settings.language === 'es' ? 'Recordatorio activo' : 'Reminder active')
                }
              </div>
            </div>

            {/* Weather Card */}
            <div className="card weather-card">
              <div className="card-header">
                <span className="card-label">
                  <span className="material-icons-outlined">thermostat</span>
                  {t.weather}
                </span>
              </div>
              
              <div className="weather-main">
                <span className="material-icons-outlined weather-icon">wb_sunny</span>
                <div>
                  <span className="weather-temp">{weather?.temp || '--'}</span>
                  <span className="weather-temp-unit">C</span>
                </div>
                <div className="weather-info">
                  <div className="weather-desc">{weather?.desc || '--'}</div>
                  <div className="weather-hydro">{weather?.hydroRec || ''}</div>
                </div>
              </div>
              
              <div className="weather-search">
                <span className="material-icons-outlined">search</span>
                <input
                  type="text"
                  placeholder={t.placeholder}
                  value={location}
                  onChange={(e) => setLocation(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && setLocationHandler()}
                />
              </div>
            </div>
          </div>
        )}

        {/* ==================== HISTORY VIEW ==================== */}
        {activeNav === 'history' && (
          <div className="history-view">
            <div className="view-header">
              <h1>{t.history}</h1>
              <div className="period-toggle">
                <button 
                  className={`period-btn ${period === 'week' ? 'active' : ''}`}
                  onClick={() => setPeriod('week')}
                >
                  {t.week}
                </button>
                <button 
                  className={`period-btn ${period === 'month' ? 'active' : ''}`}
                  onClick={() => setPeriod('month')}
                >
                  {t.month}
                </button>
              </div>
            </div>

            {/* Stats Grid */}
            <div className="stats-grid">
              <div className="stat-card">
                <span className="material-icons-outlined">functions</span>
                <div className="stat-content">
                  <span className="stat-value">{(totalMl / 1000).toFixed(1)} L</span>
                  <span className="stat-label">{t.total}</span>
                </div>
              </div>
              <div className="stat-card">
                <span className="material-icons-outlined">analytics</span>
                <div className="stat-content">
                  <span className="stat-value">{(avgMl / 1000).toFixed(1)} L</span>
                  <span className="stat-label">{t.avg}</span>
                </div>
              </div>
              <div className="stat-card">
                <span className="material-icons-outlined">emoji_events</span>
                <div className="stat-content">
                  <span className="stat-value">{goalsMet}</span>
                  <span className="stat-label">{t.goal_achieved}</span>
                </div>
              </div>
              <div className="stat-card">
                <span className="material-icons-outlined">star</span>
                <div className="stat-content">
                  <span className="stat-value">{(bestDay / 1000).toFixed(1)} L</span>
                  <span className="stat-label">{t.best_day}</span>
                </div>
              </div>
            </div>

            {/* Big Chart */}
            <div className="card chart-card">
              <div className="big-chart">
                {historyData.map((d, i) => {
                  const height = maxMl > 0 ? Math.max(8, (d.ml / maxMl) * 200) : 8;
                  const isToday = i === historyData.length - 1;
                  const isGoal = d.ml >= goal;
                  return (
                    <div key={d.date} className="bar-wrapper big">
                      <div className="bar-info">
                        <span className="bar-value">{(d.ml / 1000).toFixed(1)}L</span>
                      </div>
                      <div 
                        className={`bar ${isToday ? 'today' : isGoal ? 'goal-met' : d.ml > 0 ? 'partial' : ''}`}
                        style={{ height: `${height}px` }}
                      />
                      <span className={`bar-label ${isToday ? 'today' : ''}`}>{d.day}</span>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        )}

        {/* ==================== SETTINGS VIEW ==================== */}
        {activeNav === 'settings' && (
          <div className="settings-view">
            <div className="view-header">
              <h1>{t.settings}</h1>
            </div>

            <div className="settings-grid">
              {/* Weight Setting */}
 <div className="card setting-card">
 <div className="setting-header">
 <span className="material-icons-outlined">monitor_weight</span>
 <span className="setting-title">{t.weight}</span>
 </div>
 <div className="setting-input">
 <input
 type="number"
 value={weightInput}
 onChange={(e) => setWeightInput(e.target.value)}
 min={30}
 max={200}
 />
 <span className="setting-unit">{t.kg}</span>
 </div>
 <button className="setting-save" onClick={setWeightHandler}>
 {t.save}
 </button>
 </div>

 {/* Height Setting */}
 <div className="card setting-card">
 <div className="setting-header">
 <span className="material-icons-outlined">height</span>
 <span className="setting-title">{t.height}</span>
 </div>
 <div className="setting-input">
 <input
 type="number"
 value={heightInput}
 onChange={(e) => setHeightInput(e.target.value)}
 min={100}
 max={250}
 />
 <span className="setting-unit">{t.cm}</span>
 </div>
 <button className="setting-save" onClick={setHeightHandler}>
 {t.save}
 </button>
 </div>

 {/* Reminder Interval Setting */}
              <div className="card setting-card">
                <div className="setting-header">
                  <span className="material-icons-outlined">notifications_active</span>
                  <span className="setting-title">{t.reminder_interval}</span>
                </div>
                <div className="setting-input">
                  <input
                    type="number"
                    value={intervalInput}
                    onChange={(e) => setIntervalInput(e.target.value)}
                    min={5}
                    max={120}
                  />
                  <span className="setting-unit">{t.minutes}</span>
                </div>
                <button className="setting-save" onClick={setReminderHandler}>
                  {t.save}
                </button>
              </div>

              {/* Location Setting */}
              <div className="card setting-card">
                <div className="setting-header">
                  <span className="material-icons-outlined">place</span>
                  <span className="setting-title">{t.location_label}</span>
                </div>
                <div className="setting-input">
                  <input
                    type="text"
                    value={location}
                    onChange={(e) => setLocation(e.target.value)}
                    placeholder={t.placeholder}
                  />
                </div>
                <button className="setting-save" onClick={setLocationHandler}>
                  {t.save}
                </button>
              </div>

              {/* Daily Goal Display */}
              <div className="card setting-card info-card">
                <div className="setting-header">
                  <span className="material-icons-outlined">flag</span>
                  <span className="setting-title">{t.daily_progress}</span>
                </div>
                <div className="goal-display">
                  <span className="goal-value">{(goal / 1000).toFixed(2)} L</span>
                  <span className="goal-hint">({settings.language === 'es' ? 'basado en peso y estatura' : 'based on weight and height'})</span>
                </div>
              </div>
            </div>
          </div>
        )}
      </main>

      {/* Reminder Modal */}
      {showReminder && reminderData && (
        <ReminderModal
          consumed={reminderData.consumed}
          goal={reminderData.goal}
          onClose={() => setShowReminder(false)}
          onAdd={() => {
            loadSettings();
            loadHistory();
          }}
        />
      )}
    </div>
  );
}

export default App;
