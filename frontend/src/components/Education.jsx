export default function Education({ items, photoUrl }) {
  return (
    <section id="education" className="section education"
      style={photoUrl ? { '--college-photo': `url(${JSON.stringify(photoUrl)})` } : undefined}>
      <div className="container">
        <h2>Education</h2>
        <div className="timeline">
          {items.map((e) => (
            <div className="timeline-item" key={e.id}>
              <div className="timeline-date">{e.startYear} – {e.endYear}</div>
              <h3>{e.degree}{e.fieldOfStudy ? `, ${e.fieldOfStudy}` : ''}</h3>
              <div className="timeline-org">{e.institution}</div>
              {e.description && <p>{e.description}</p>}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
