import { useState, useEffect } from 'react';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { GetSettings, AddWater, SetWeight, SetLanguage, SetLocation, SetReminderInterval, GetWeeklyData, GetMonthlyData, SetMood, GetMoodRecommendation } from '../wailsjs/go/main/App';
import ReminderModal from './ReminderModal';
import './App.css';

interface Settings {
  weight: number;
  today_consumed: number;
  daily_goal: number;
  language: string;
  location: string;
  last_reset_date: string;
  reminder_interval: number;
  current_mood: string;
}

interface HistoryDay {
  day: string;
  ml: number;
  date: string;
}

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
    daily_progress: "PROGRESO DIARIO",
    register_intake: "REGISTRAR INGESTA",
    mood: "ESTADO MENTAL",
    history: "HISTORIAL",
    weather: "CONDICIONES AMBIENTALES",
    settings: "CONFIGURACION",
    week: "Semana",
    month: "Mes",
    avg: "Promedio",
    total: "Total",
    placeholder: "Escribe tu ciudad...",
    mood_well: "Bien",
    mood_neutral: "Neutral",
    mood_low: "Bajo",
    mood_tense: "Tenso",
  },
  en: {
    daily_progress: "DAILY PROGRESS",
    register_intake: "LOG INTAKE",
    mood: "MENTAL STATE",
    history: "HISTORY",
    weather: "WEATHER CONDITIONS",
    settings: "SETTINGS",
    week: "Week",
    month: "Month",
    avg: "Average",
    total: "Total",
    placeholder: "Enter your city...",
    mood_well: "Well",
    mood_neutral: "Neutral",
    mood_low: "Low",
    mood_tense: "Tense",
  },
};

const moodEmojis: Record<string, { emoji: string; color: string }> = {
  well: { emoji: "😊", color: "#4ade80" },
  neutral: { emoji: "😐", color: "#94a3b8" },
  low: { emoji: "😔", color: "#60a5fa" },
  tense: { emoji: "😤", color: "#f87171" },
};

function App() {
  const [settings, setSettings] = useState<Settings>({
    weight: 70,
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

  const t = translations[settings.language as "es" | "en"] || translations.es;
  const consumed = settings.today_consumed;
  const goal = settings.daily_goal;
  const progress = Math.min(100, (consumed / goal) * 100);
  const remaining = Math.max(0, goal - consumed);
  const segmentsFilled = Math.min(10, Math.floor((consumed / goal) * 10));

 useEffect(() => {
 loadSettings();
 loadHistory();

 // Escuchar evento de recordatorio desde Go
 EventsOn('show-reminder', (data: { consumed: number; goal: number }) => {
 setReminderData(data);
 setShowReminder(true);
 });
 }, []);

  useEffect(() => {
    if (settings.location && !location) {
      setLocation(settings.location);
    }
  }, [settings.location]);

  useEffect(() => {
    if (settings.location) {
      fetchWeather(settings.location);
    }
  }, [settings.location]);

  // Cargar recomendacion de mood cuando cambie
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

      if (!geoData.results || geoData.results.length === 0) {
        console.error("Ciudad no encontrada");
        return;
      }

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
    if (temp >= 30) {
      return lang === "es" ? "Bebe mas agua de lo habitual" : "Drink more water than usual";
    } else if (temp >= 25) {
      return lang === "es" ? "Mantente hidratado" : "Stay hydrated";
    } else if (temp <= 10) {
      return lang === "es" ? "Agua templada ayuda" : "Warm water helps";
    }
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
      const data = period === "week"
        ? await GetWeeklyData()
        : await GetMonthlyData();
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

  const setWeight = async (weight: number) => {
    try {
      const s = await SetWeight(weight);
      setSettings(s);
    } catch (e) {
      console.error("Failed to set weight:", e);
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

  const setReminder = async (interval: number) => {
    try {
      const s = await SetReminderInterval(interval);
      setSettings(s);
    } catch (e) {
      console.error("Failed to set reminder:", e);
    }
  };

  const setLocationHandler = async () => {
    try {
      const s = await SetLocation(location);
      setSettings(s);
      if (location) {
        fetchWeather(location);
      }
    } catch (e) {
      console.error("Failed to set location:", e);
    }
  };

  const maxMl = Math.max(...historyData.map(d => d.ml), goal);
  const totalMl = historyData.reduce((sum, d) => sum + d.ml, 0);
  const avgMl = historyData.length > 0 ? Math.round(totalMl / historyData.length) : 0;

  const currentMoodData = moodEmojis[settings.current_mood] || moodEmojis.neutral;

  return (
    <div className="app">
      <div className="header">
        <span className="header-title">Hydrapotion</span>
      </div>

      {/* Progress Card */}
      <div className="card">
        <div className="card-label">{t.daily_progress}</div>
        <div className="progress-top">
          <div style={{ display: "flex", alignItems: "flex-end" }}>
            <span className="big-num">{(consumed / 1000).toFixed(1)}</span>
            <span className="big-unit">L</span>
          </div>
          <span className="goal-badge">meta {(goal / 1000).toFixed(1)} L</span>
        </div>
        <div className="pbar-track">
          <div className="pbar-fill" style={{ width: `${progress}%` }} />
        </div>
        <div className="pbar-info">
          <span>{Math.round(progress)}% completado</span>
          <span>faltan {(remaining / 1000).toFixed(1)} L</span>
        </div>
        <div className="segments">
          {[...Array(10)].map((_, i) => (
            <div key={i} className={`seg ${i < segmentsFilled ? "on" : ""}`} />
          ))}
        </div>
      </div>

      {/* Mood Card */}
      <div className="card">
        <div className="card-label">{t.mood}</div>
        <div className="mood-selector">
          {Object.entries(moodEmojis).map(([mood, data]) => (
            <button
              key={mood}
              className={`mood-btn ${settings.current_mood === mood ? "active" : ""}`}
              onClick={() => setMoodHandler(mood)}
              style={{ "--mood-color": data.color } as React.CSSProperties}
            >
              <span className="mood-emoji">{data.emoji}</span>
              <span className="mood-label">
                {t[`mood_${mood}` as keyof typeof t]}
              </span>
            </button>
          ))}
        </div>
        {moodRec && (
          <div className="mood-rec">
            <span className="mood-rec-text">{moodRec.recommendation}</span>
            {moodRec.multiplier > 1 && (
              <span className="mood-rec-adjust">{moodRec.adjustment}</span>
            )}
          </div>
        )}
      </div>

      {/* Ingesta Card */}
      <div className="card">
        <div className="card-label">{t.register_intake}</div>
        <div className="btn-row">
          <button className="qty-btn" onClick={() => addWater(150)}>
            <span className="ml">150</span>
            <span className="sub">ml</span>
          </button>
          <button className="qty-btn primary" onClick={() => addWater(250)}>
            <span className="ml">250</span>
            <span className="sub">ml</span>
          </button>
          <button className="qty-btn" onClick={() => addWater(500)}>
            <span className="ml">500</span>
            <span className="sub">ml</span>
          </button>
        </div>
      </div>

      {/* History Card */}
      <div className="card">
        <div className="hist-header">
          <span className="card-label" style={{ marginBottom: 0 }}>{t.history}</span>
          <div className="period-pills">
            <span
              className={`period-pill ${period === "week" ? "on" : ""}`}
              onClick={() => setPeriod("week")}
            >
              {t.week}
            </span>
            <span
              className={`period-pill ${period === "month" ? "on" : ""}`}
              onClick={() => setPeriod("month")}
            >
              {t.month}
            </span>
          </div>
        </div>
        <div className="bars">
          {historyData.map((d, i) => {
            const height = maxMl > 0 ? (d.ml / maxMl) * 50 : 2;
            const isToday = i === historyData.length - 1;
            const isGoal = d.ml >= goal;
            return (
              <div key={d.date} className="bar-wrap">
                <div
                  className={`bar ${isToday ? "today" : isGoal ? "goal" : d.ml > 0 ? "partial" : ""}`}
                  style={{ height: `${height}px` }}
                />
                <span className={`bar-day ${isToday ? "now" : ""}`}>{d.day}</span>
              </div>
            );
          })}
        </div>
        <div className="hist-stats">
          <span>{t.avg} <b>{(avgMl / 1000).toFixed(1)} L</b></span>
          <span>{t.total} <b>{(totalMl / 1000).toFixed(1)} L</b></span>
        </div>
      </div>

      {/* Weather Card */}
      <div className="card">
        <div className="card-label">{t.weather}</div>
        <div className="weather-row">
          <span className="temp-big">
            {weather ? `${weather.temp}°` : '--°'}
            <span style={{ fontSize: "16px", color: "var(--muted)" }}>C</span>
          </span>
          <div>
            <div className="weather-desc">{weather?.desc || '--'}</div>
            <div className="hydro-rec">{weather?.hydroRec || ''}</div>
          </div>
        </div>
        <div className="weather-input">
          <div className="search-icon" />
          <input
            type="text"
            placeholder={t.placeholder}
            value={location}
            onChange={(e) => setLocation(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && setLocationHandler()}
          />
        </div>
      </div>

      {/* Settings Card */}
      <div className="card">
        <div className="card-label">{t.settings}</div>
        <div className="cfg-row">
          <span className="cfg-label">Peso</span>
          <span
            className="cfg-val"
            onClick={() => {
              const newWeight = prompt("Ingresa tu peso (kg):", String(settings.weight));
              if (newWeight) setWeight(parseInt(newWeight));
            }}
          >
            {settings.weight} kg
          </span>
        </div>
        <div className="cfg-row">
          <span className="cfg-label">Recordatorio</span>
          <span className="cfg-val">
            <select
              value={settings.reminder_interval}
              onChange={(e) => setReminder(parseInt(e.target.value))}
            >
              <option value={900}>15 min</option>
              <option value={1800}>30 min</option>
              <option value={3600}>1 hora</option>
              <option value={7200}>2 horas</option>
            </select>
          </span>
        </div>
        <div className="cfg-row">
          <span className="cfg-label">Idioma</span>
          <span className="cfg-val">
            <select
              value={settings.language}
              onChange={(e) => setLanguage(e.target.value)}
            >
 <option value="es">Espanol</option>
 <option value="en">English</option>
 </select>
 </span>
 </div>
 </div>

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
