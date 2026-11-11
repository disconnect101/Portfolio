import { useEffect, useState } from 'react';
import Navbar from './components/Navbar.jsx';
import Hero from './components/Hero.jsx';
import Projects from './components/Projects.jsx';
import Experience from './components/Experience.jsx';
import Education from './components/Education.jsx';
import Footer from './components/Footer.jsx';
import Admin from './components/Admin.jsx';

// Set VITE_API_URL (e.g. https://api.example.com) when the API is on a different origin.
const API = import.meta.env.VITE_API_URL || '';

function useTheme() {
  const [theme, setTheme] = useState(() => {
    try {
      const saved = localStorage.getItem('theme');
      if (saved) return saved;
    } catch { /* ignore */ }
    return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  });

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    try { localStorage.setItem('theme', theme); } catch { /* ignore */ }
  }, [theme]);

  return [theme, () => setTheme((t) => (t === 'dark' ? 'light' : 'dark'))];
}

export default function App() {
  if (window.location.pathname === '/admin' || window.location.pathname === '/admin/') {
    return <Admin />;
  }
  return <Portfolio />;
}

function Portfolio() {
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);
  const [theme, toggleTheme] = useTheme();

  useEffect(() => {
    fetch(`${API}/api/portfolio`)
      .then((r) => {
        if (!r.ok) throw new Error(`Request failed (${r.status})`);
        return r.json();
      })
      .then((d) => {
        setData(d);
        document.title = `${d.profile.name} | Portfolio`;
      })
      .catch((e) => setError(e.message));
  }, []);

  if (error) {
    return (
      <div className="status">
        <h2>Couldn't load the portfolio</h2>
        <p>{error}. Make sure the backend is running on port 8080.</p>
      </div>
    );
  }
  if (!data) return <div className="status">Loading…</div>;

  const { profile, projects, education, experience } = data;

  return (
    <>
      <Navbar name={profile.name} theme={theme} onToggleTheme={toggleTheme} />
      <main>
        <Hero profile={profile} />
        <Experience items={experience} />
        <Education items={education} photoUrl={data.educationPhotoPath && API + data.educationPhotoPath} />
        <Projects projects={projects} />
      </main>
      <Footer profile={profile} />
    </>
  );
}
