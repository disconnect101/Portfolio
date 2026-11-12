import { useEffect, useState } from 'react';
import CompanyLogo from './CompanyLogo.jsx';

const API = import.meta.env.VITE_API_URL || '';

const sections = [
  {
    key: 'projects', label: 'Projects', required: 'title',
    fields: [
      ['title', 'Title'], ['description', 'Description', 'textarea'],
      ['technologies', 'Technologies (one per line)', 'list'],
      ['imageUrl', 'Image URL'], ['liveUrl', 'Live URL'], ['repoUrl', 'Repository URL'],
      ['displayOrder', 'Display order', 'number'],
    ],
  },
  {
    key: 'education', label: 'Education', required: 'institution',
    fields: [
      ['institution', 'Institution'], ['degree', 'Degree'], ['fieldOfStudy', 'Field of study'],
      ['startYear', 'Start year'], ['endYear', 'End year'],
      ['description', 'Description', 'textarea'], ['displayOrder', 'Display order', 'number'],
    ],
  },
  {
    key: 'experience', label: 'Experience', required: 'company',
    fields: [
      ['company', 'Company'], ['role', 'Role'], ['location', 'Location'],
      ['startDate', 'Start date'], ['endDate', 'End date'],
      ['highlights', 'Highlights (one per line)', 'list'],
      ['displayOrder', 'Display order', 'number'],
    ],
  },
];

const profileFields = [
  ['name', 'Name'], ['title', 'Title'], ['description', 'About', 'textarea'],
  ['coverPhotoUrl', 'Cover photo URL'], ['profilePhotoUrl', 'Profile photo URL'],
  ['email', 'Email', 'email'], ['location', 'Location'],
  ['githubUrl', 'GitHub URL'], ['linkedinUrl', 'LinkedIn URL'],
];

async function request(path, csrfToken, options = {}) {
  const response = await fetch(`${API}${path}`, {
    ...options,
    cache: 'no-store',
    headers: {
      ...(csrfToken ? { 'X-CSRF-Token': csrfToken } : {}),
      ...(options.body && !(options.body instanceof File) ? { 'Content-Type': 'application/json' } : {}),
    },
  });
  if (!response.ok) {
    if (response.status === 401) throw new Error('Invalid key or expired session. Please log in again.');
    if (response.status === 403) throw new Error('Session security check failed. Please refresh and log in again.');
    if (response.status === 503) throw new Error('Admin account is not configured on the server.');
    if (options.body instanceof File && [400, 413, 415].includes(response.status)) {
      const message = response.headers.get('Content-Type')?.startsWith('text/plain') && await response.text();
      throw new Error(message || 'Check the file type and upload size.');
    }
    throw new Error(`Request failed (${response.status}). Check the fields and try again.`);
  }
  return response.status === 204 ? null : response.json();
}

function Field({ field, value, onChange, required }) {
  const [name, label, type] = field;
  const display = type === 'list' ? (value || []).join('\n') : (value ?? '');
  const update = (event) => {
    const text = event.target.value;
    onChange(name, type === 'list' ? text.split('\n').map((x) => x.trim()).filter(Boolean)
      : type === 'number' ? Number(text) : text);
  };
  return (
    <label className="admin-field">
      <span>{label}</span>
      {type === 'textarea' || type === 'list'
        ? <textarea value={display} onChange={update} rows={type === 'list' ? 3 : 5} required={required} />
        : <input type={type === 'number' || type === 'email' ? type : 'text'} value={display}
            onChange={update} required={required} />}
    </label>
  );
}

function UploadForm({ title, buttonLabel, children, onUpload,
  accept = 'image/jpeg,image/png', fileLabel = 'Choose a JPEG or PNG (up to 5 MiB and 24 megapixels)' }) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [saved, setSaved] = useState(false);

  async function upload(event) {
    event.preventDefault();
    const form = event.currentTarget;
    const file = form.elements.upload.files[0];
    if (!file) return;
    setError('');
    setSaved(false);
    if (file.size > 5 * 1024 * 1024) {
      setError('Choose a file of 5 MiB or less.');
      return;
    }
    setBusy(true);
    try {
      await onUpload(file);
      form.reset();
      setSaved(true);
    } catch (e) {
      setError(e.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <form className="admin-card admin-photo" onSubmit={upload}>
      <h3>{title}</h3>
      {children}
      <label className="admin-field">{fileLabel}
        <input type="file" name="upload" accept={accept} required disabled={busy}
          onChange={() => { setSaved(false); setError(''); }} />
      </label>
      {error && <p className="admin-error" role="alert">{error}</p>}
      {saved && <p className="admin-success" role="status">Upload saved.</p>}
      <div className="admin-actions">
        <button className="btn primary" disabled={busy}>{busy ? 'Uploading…' : buttonLabel}</button>
      </div>
    </form>
  );
}

function Editor({ title, fields, required, initial, onSave, onDelete }) {
  const [values, setValues] = useState(initial);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [saved, setSaved] = useState(false);

  async function save(event) {
    event.preventDefault();
    setBusy(true);
    setError('');
    setSaved(false);
    try {
      await onSave(values);
      setSaved(true);
      if (!initial.id) setValues(initial);
    } catch (e) {
      setError(e.message);
    } finally {
      setBusy(false);
    }
  }

  async function remove() {
    if (!window.confirm(`Delete ${title}? This cannot be undone.`)) return;
    setBusy(true);
    setError('');
    try {
      await onDelete();
    } catch (e) {
      setError(e.message);
      setBusy(false);
    }
  }

  return (
    <form className="admin-card" onSubmit={save}>
      <h3>{title}</h3>
      <div className="admin-fields">
        {fields.map((field) => <Field key={field[0]} field={field} value={values[field[0]]}
          required={field[0] === required}
          onChange={(name, value) => { setValues((old) => ({ ...old, [name]: value })); setSaved(false); }} />)}
      </div>
      {error && <p className="admin-error" role="alert">{error}</p>}
      {saved && <p className="admin-success" role="status">Saved.</p>}
      <div className="admin-actions">
        <button className="btn primary" type="submit" disabled={busy}>{busy ? 'Working…' : 'Save'}</button>
        {onDelete && <button className="btn admin-delete" type="button" onClick={remove} disabled={busy}>Delete</button>}
      </div>
    </form>
  );
}

export default function Admin() {
  const [inputKey, setInputKey] = useState('');
  const [session, setSession] = useState(null);
  const [data, setData] = useState(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);

  async function load() {
    const response = await fetch(`${API}/api/portfolio`, { cache: 'no-store' });
    if (!response.ok) throw new Error(`Couldn't load portfolio (${response.status}).`);
    setData(await response.json());
  }

  async function uploadFile(path, file) {
    await request(path, session.csrfToken, { method: 'PUT', body: file });
    await load();
  }

  useEffect(() => {
    request('/api/admin/session', '')
      .then(async (savedSession) => { await load(); setSession(savedSession); })
      .catch(() => { /* no active session: show login */ })
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    if (!session) return undefined;
    const timer = window.setTimeout(() => { setSession(null); setData(null); },
      Math.max(0, session.expiresAt - Date.now()));
    return () => window.clearTimeout(timer);
  }, [session]);

  async function login(event) {
    event.preventDefault();
    if (window.location.protocol !== 'https:' && !['localhost', '127.0.0.1'].includes(window.location.hostname)) {
      setError('Open the HTTPS version of this site before entering your admin key.');
      return;
    }
    setLoading(true);
    setError('');
    try {
      const authenticated = await request('/api/admin/login', '',
        { method: 'POST', body: JSON.stringify({ key: inputKey }) });
      await load();
      setSession(authenticated);
      setInputKey('');
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }

  async function logout() {
    try { await request('/api/admin/logout', session.csrfToken, { method: 'POST' }); }
    catch { /* clear local state even if the server session already expired */ }
    setSession(null);
    setData(null);
    setInputKey('');
    setError('');
  }

  if (loading && !session) return <div className="status">Checking admin session…</div>;

  if (!session) {
    return (
      <main className="admin-login">
        <form className="admin-card" onSubmit={login}>
          <h1>Portfolio admin</h1>
          <p>Enter your admin key to start a time-limited session.</p>
          <label className="admin-field"><span>Admin key</span>
            <input type="password" autoComplete="off" value={inputKey}
              onChange={(e) => setInputKey(e.target.value)} required /></label>
          {error && <p className="admin-error" role="alert">{error}</p>}
          <button className="btn primary" type="submit" disabled={loading}>{loading ? 'Checking…' : 'Unlock'}</button>
          <a href="/">Back to portfolio</a>
        </form>
      </main>
    );
  }

  return (
    <main className="admin-page container">
      <header className="admin-header">
        <div><h1>Portfolio admin</h1><p>Changes appear on the public site after you save.</p></div>
        <div className="admin-actions"><a className="btn" href="/">View site</a><button className="btn" onClick={logout}>Log out</button></div>
      </header>
      <section className="admin-section">
        <h2>Profile</h2>
        {Object.hasOwn(data.profile, 'hasResume') && (
          <UploadForm title="Resume" buttonLabel="Upload Resume" accept="application/pdf,.pdf"
            fileLabel="Choose a PDF (up to 5 MiB)" onUpload={(file) => uploadFile('/api/admin/resume', file)}>
            {data.profile.hasResume && <p><a href={API + '/api/resume'} target="_blank" rel="noreferrer">View current Resume</a></p>}
          </UploadForm>
        )}
        <Editor key={`profile-${data.profile?.id}`} title="Your profile" fields={profileFields}
          required="name" initial={data.profile || {}}
          onSave={async (values) => { await request('/api/admin/profile', session.csrfToken,
            { method: 'PUT', body: JSON.stringify(values) }); await load(); }} />
      </section>
      {sections.map((section) => (
        <section className="admin-section" key={section.key}>
          <h2>{section.label}</h2>
          {section.key === 'education' && Object.hasOwn(data, 'educationPhotoPath') &&
            <UploadForm title="College background photo" buttonLabel="Upload photo"
              onUpload={(file) => uploadFile('/api/admin/education-photo', file)}>
              <img src={data.educationPhotoPath ? API + data.educationPhotoPath : '/college-campus.jpg'} alt="Current Education background" />
            </UploadForm>}
          <div className="admin-list">
            {data[section.key].map((item) => (
              <div key={item.id}>
                <Editor title={item[section.required] || `${section.label} #${item.id}`}
                  fields={section.fields} required={section.required} initial={item}
                  onSave={async (values) => { await request(`/api/admin/${section.key}/${item.id}`, session.csrfToken,
                    { method: 'PUT', body: JSON.stringify(values) }); await load(); }}
                  onDelete={async () => { await request(`/api/admin/${section.key}/${item.id}`, session.csrfToken,
                    { method: 'DELETE' }); await load(); }} />
                {section.key === 'experience' && item.company?.trim() && Object.hasOwn(item, 'companyLogoPath') &&
                  <UploadForm key={item.company} title={`Logo for ${item.company}`} buttonLabel="Upload logo"
                    onUpload={(file) => uploadFile(`/api/admin/experience/${item.id}/logo`, file)}>
                    <CompanyLogo company={item.company} logoPath={item.companyLogoPath} />
                  </UploadForm>}
              </div>
            ))}
            <Editor key={`new-${section.key}`} title={`Add ${section.label.toLowerCase()}`}
              fields={section.fields} required={section.required} initial={{ displayOrder: data[section.key].length + 1 }}
              onSave={async (values) => { await request(`/api/admin/${section.key}`, session.csrfToken,
                { method: 'POST', body: JSON.stringify(values) }); await load(); }} />
          </div>
        </section>
      ))}
    </main>
  );
}
