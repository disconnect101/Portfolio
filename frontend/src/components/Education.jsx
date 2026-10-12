export default function Education({ items }) {
  return (
    <section id="education" className="section alt">
      <div className="container narrow">
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
