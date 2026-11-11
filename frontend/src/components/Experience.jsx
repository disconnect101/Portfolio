import CompanyLogo, { companyLogoSource } from './CompanyLogo.jsx';

export default function Experience({ items }) {
  const logos = [...new Set(items.map(e => companyLogoSource(e.company, e.companyLogoPath)).filter(Boolean))];
  return (
    <section id="experience" className="section">
      <div className="container">
        <h2>Experience</h2>
        <div className={logos.length ? 'experience-layout' : undefined}>
          <div className="timeline">
            {items.map((e) => (
              <div className="timeline-item" key={e.id}>
                <CompanyLogo company={e.company} logoPath={e.companyLogoPath} />
                <h3>{e.role}</h3>
                <div className="timeline-date">{e.startDate} – {e.endDate}</div>
                <div className="timeline-org">{e.company}{e.location ? ` · ${e.location}` : ''}</div>
                <ul>
                  {e.highlights.map((h, i) => <li key={i}>{h}</li>)}
                </ul>
              </div>
            ))}
          </div>
          {logos.length > 0 && (
            <div className="experience-logo-grid" aria-hidden="true">
              {logos.map(src => <img key={src} src={src} alt="" />)}
            </div>
          )}
        </div>
      </div>
    </section>
  );
}
